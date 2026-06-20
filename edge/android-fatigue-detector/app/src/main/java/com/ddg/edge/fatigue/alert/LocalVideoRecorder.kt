package com.ddg.edge.fatigue.alert

import android.content.Context
import android.media.MediaCodec
import android.media.MediaCodecInfo
import android.media.MediaFormat
import android.media.MediaMuxer
import android.os.Environment
import android.view.Surface
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.withContext
import java.io.File
import java.nio.ByteBuffer
import java.text.SimpleDateFormat
import java.util.Date
import java.util.LinkedList
import java.util.Locale

class LocalVideoRecorder(
    private val context: Context,
    private val width: Int = 640,
    private val height: Int = 480,
    private val frameRate: Int = 30,
    private val bitRate: Int = 2_000_000,
    private val preAlarmBufferSeconds: Int = 10,
    private val postAlarmRecordSeconds: Int = 5
) {

    private var encoder: MediaCodec? = null
    private var muxer: MediaMuxer? = null
    private var inputSurface: Surface? = null
    private var trackIndex: Int = -1
    private var isRecording = false
    private var isMuxerStarted = false

    private val bufferDurationMs = preAlarmBufferSeconds * 1000L
    private val frameDurationUs = 1_000_000L / frameRate

    private data class BufferFrame(
        val byteBuffer: ByteBuffer,
        val bufferInfo: MediaCodec.BufferInfo,
        val timestampUs: Long
    )

    private val frameBuffer = LinkedList<BufferFrame>()

    private var outputDir: File? = null
    private var currentOutputFile: File? = null
    private var alarmTriggeredAt: Long? = null
    private var recordingStartTime: Long? = null
    private var presentationStartTimeUs: Long = 0

    private val _recordingEvents = Channel<RecordingEvent>(Channel.UNLIMITED)
    val recordingEvents = _recordingEvents

    sealed class RecordingEvent {
        data class Started(val file: File) : RecordingEvent()
        data class Saved(val file: File, val durationMs: Long) : RecordingEvent()
        data class Error(val message: String) : RecordingEvent()
    }

    @Synchronized
    suspend fun initialize(): Surface? = withContext(Dispatchers.Default) {
        runCatching {
            ensureOutputDir()
            setupEncoder()
            inputSurface
        }.onFailure {
            _recordingEvents.trySend(RecordingEvent.Error("Init failed: ${it.message}"))
        }.getOrNull()
    }

    private fun setupEncoder() {
        val format = MediaFormat.createVideoFormat(MediaFormat.MIMETYPE_VIDEO_AVC, width, height).apply {
            setInteger(MediaFormat.KEY_COLOR_FORMAT, MediaCodecInfo.CodecCapabilities.COLOR_FormatSurface)
            setInteger(MediaFormat.KEY_BIT_RATE, bitRate)
            setInteger(MediaFormat.KEY_FRAME_RATE, frameRate)
            setInteger(MediaFormat.KEY_I_FRAME_INTERVAL, 1)
            setInteger(MediaFormat.KEY_PROFILE, MediaCodecInfo.CodecProfileLevel.AVCProfileBaseline)
            setInteger(MediaFormat.KEY_LEVEL, MediaCodecInfo.CodecProfileLevel.AVCLevel31)
        }

        encoder = MediaCodec.createEncoderByType(MediaFormat.MIMETYPE_VIDEO_AVC).also { codec ->
            codec.configure(format, null, null, MediaCodec.CONFIGURE_FLAG_ENCODE)
            inputSurface = codec.createInputSurface()
            codec.start()
        }

        startDrainingThread()
    }

    private fun ensureOutputDir() {
        val dir = File(
            context.getExternalFilesDir(Environment.DIRECTORY_MOVIES),
            "alarm_videos"
        )
        if (!dir.exists()) {
            dir.mkdirs()
        }
        outputDir = dir
    }

    private fun startDrainingThread() {
        Thread {
            val bufferInfo = MediaCodec.BufferInfo()
            while (encoder != null) {
                try {
                    val status = encoder!!.dequeueOutputBuffer(bufferInfo, 10_000)
                    when {
                        status == MediaCodec.INFO_OUTPUT_FORMAT_CHANGED -> {
                            handleFormatChange()
                        }
                        status >= 0 -> {
                            handleOutputBuffer(status, bufferInfo)
                        }
                    }
                } catch (e: IllegalStateException) {
                    break
                }
            }
        }.start()
    }

    private fun handleFormatChange() {
        val newFormat = encoder!!.outputFormat
        synchronized(this) {
            if (muxer != null && !isMuxerStarted) {
                trackIndex = muxer!!.addTrack(newFormat)
                muxer!!.start()
                isMuxerStarted = true
            }
        }
    }

    @Synchronized
    private fun handleOutputBuffer(index: Int, info: MediaCodec.BufferInfo) {
        val encoderRef = encoder ?: return
        val outputBuffer = encoderRef.getOutputBuffer(index) ?: return

        val adjustedBuffer = outputBuffer.duplicate()
        adjustedBuffer.position(info.offset)
        adjustedBuffer.limit(info.offset + info.size)

        val copyBuffer = ByteBuffer.allocate(info.size)
        copyBuffer.put(adjustedBuffer)
        copyBuffer.flip()

        val copyInfo = MediaCodec.BufferInfo().apply {
            set(info.offset, info.size, info.presentationTimeUs, info.flags)
        }

        if (isRecording && alarmTriggeredAt == null) {
            addToBuffer(copyBuffer, copyInfo, info.presentationTimeUs)
        } else if (isRecording && alarmTriggeredAt != null) {
            writeBufferedFrames()
            writeFrameToMuxer(copyBuffer, copyInfo)
            checkPostAlarmDuration(info.presentationTimeUs)
        } else {
            addToBuffer(copyBuffer, copyInfo, info.presentationTimeUs)
        }

        encoderRef.releaseOutputBuffer(index, false)
    }

    private fun addToBuffer(
        buffer: ByteBuffer,
        info: MediaCodec.BufferInfo,
        timestampUs: Long
    ) {
        frameBuffer.addLast(BufferFrame(buffer, info, timestampUs))

        while (frameBuffer.size > 2) {
            val first = frameBuffer.first
            val last = frameBuffer.last
            if (last.timestampUs - first.timestampUs > bufferDurationMs * 1000) {
                frameBuffer.removeFirst()
            } else {
                break
            }
        }
    }

    @Synchronized
    private fun writeBufferedFrames() {
        while (frameBuffer.isNotEmpty()) {
            val frame = frameBuffer.removeFirst()
            val adjustedInfo = MediaCodec.BufferInfo().apply {
                set(
                    frame.bufferInfo.offset,
                    frame.bufferInfo.size,
                    frame.timestampUs - presentationStartTimeUs,
                    frame.bufferInfo.flags
                )
            }
            writeFrameToMuxer(frame.byteBuffer, adjustedInfo)
        }
    }

    private fun writeFrameToMuxer(buffer: ByteBuffer, info: MediaCodec.BufferInfo) {
        if (muxer != null && isMuxerStarted && trackIndex >= 0) {
            try {
                muxer!!.writeSampleData(trackIndex, buffer, info)
            } catch (e: Exception) {
            }
        }
    }

    private fun checkPostAlarmDuration(currentTimestampUs: Long) {
        val alarmTime = alarmTriggeredAt ?: return
        val postDurationUs = postAlarmRecordSeconds * 1_000_000L
        val startUs = (alarmTime * 1000L) - (preAlarmBufferSeconds * 1_000_000L)

        if (currentTimestampUs - startUs >= (preAlarmBufferSeconds + postAlarmRecordSeconds) * 1_000_000L) {
            kotlinx.coroutines.runBlocking {
                stopRecording()
            }
        }
    }

    @Synchronized
    suspend fun startRecording(): File? = withContext(Dispatchers.IO) {
        if (isRecording) return@withContext currentOutputFile

        runCatching {
            val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.getDefault()).format(Date())
            val fileName = "alarm_$timestamp.mp4"
            currentOutputFile = File(outputDir, fileName)

            muxer = MediaMuxer(
                currentOutputFile!!.absolutePath,
                MediaMuxer.OutputFormat.MUXER_OUTPUT_MPEG_4
            )
            isMuxerStarted = false
            trackIndex = -1

            isRecording = true
            recordingStartTime = System.currentTimeMillis()
            alarmTriggeredAt = null
            presentationStartTimeUs = (recordingStartTime!! - preAlarmBufferSeconds * 1000L) * 1000L

            _recordingEvents.send(RecordingEvent.Started(currentOutputFile!!))
            currentOutputFile
        }.onFailure {
            _recordingEvents.trySend(RecordingEvent.Error("Start failed: ${it.message}"))
        }.getOrNull()
    }

    @Synchronized
    fun triggerAlarm() {
        if (isRecording && alarmTriggeredAt == null) {
            alarmTriggeredAt = System.currentTimeMillis()
        }
    }

    @Synchronized
    suspend fun stopRecording(): File? = withContext(Dispatchers.IO) {
        if (!isRecording) return@withContext null

        val file = currentOutputFile
        val duration = recordingStartTime?.let { System.currentTimeMillis() - it } ?: 0L

        runCatching {
            isRecording = false
            alarmTriggeredAt = null

            if (muxer != null && isMuxerStarted) {
                muxer!!.stop()
            }
            muxer?.release()
            muxer = null
            isMuxerStarted = false
            trackIndex = -1

            currentOutputFile = null
            recordingStartTime = null

            if (file != null && file.exists()) {
                _recordingEvents.send(RecordingEvent.Saved(file, duration))
            }
            file
        }.onFailure {
            _recordingEvents.trySend(RecordingEvent.Error("Stop failed: ${it.message}"))
        }.getOrNull()
    }

    @Synchronized
    suspend fun release() = withContext(Dispatchers.Default) {
        stopRecording()
        runCatching {
            inputSurface?.release()
            encoder?.stop()
            encoder?.release()
        }
        inputSurface = null
        encoder = null
        frameBuffer.clear()
    }

    fun getInputSurface(): Surface? = inputSurface

    fun isRecordingActive(): Boolean = isRecording

    fun getOutputDirectory(): File? = outputDir
}
