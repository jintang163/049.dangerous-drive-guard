package com.ddg.edge.fatigue.utils

import android.graphics.Bitmap
import android.graphics.Matrix
import com.google.mediapipe.framework.image.MPImage
import com.google.mediapipe.framework.image.MPImage.MEDIAPIPE_IMAGE_FORMAT_RGBA
import java.nio.ByteBuffer

data class FrameMetadata(
    val frameId: Long,
    val timestamp: Long,
    val timestampNs: Long,
    val width: Int,
    val height: Int,
    val rotationDegrees: Int,
    val isMirrored: Boolean = true,
    val format: Int = FORMAT_RGBA
) {
    companion object {
        const val FORMAT_RGBA = 1
        const val FORMAT_YUV = 2
        const val FORMAT_JPEG = 3

        fun create(
            frameId: Long,
            width: Int,
            height: Int,
            rotationDegrees: Int = 0,
            isMirrored: Boolean = true
        ): FrameMetadata {
            val now = System.currentTimeMillis()
            return FrameMetadata(
                frameId = frameId,
                timestamp = now,
                timestampNs = now * 1_000_000L,
                width = width,
                height = height,
                rotationDegrees = rotationDegrees,
                isMirrored = isMirrored
            )
        }
    }

    fun isRotated(): Boolean = rotationDegrees != 0 && rotationDegrees != 180

    fun getRotatedDimensions(): Pair<Int, Int> {
        return if (isRotated()) height to width else width to height
    }
}

object FrameUtils {

    fun bitmapToMPImage(bitmap: Bitmap): MPImage {
        val width = bitmap.width
        val height = bitmap.height
        val pixelCount = width * height
        val byteBuffer = ByteBuffer.allocateDirect(pixelCount * 4)

        val pixels = IntArray(pixelCount)
        bitmap.getPixels(pixels, 0, width, 0, 0, width, height)

        for (i in pixels.indices) {
            val pixel = pixels[i]
            byteBuffer.put(((pixel shr 16) and 0xFF).toByte())
            byteBuffer.put(((pixel shr 8) and 0xFF).toByte())
            byteBuffer.put((pixel and 0xFF).toByte())
            byteBuffer.put(((pixel shr 24) and 0xFF).toByte())
        }

        byteBuffer.rewind()
        return MPImage.createFromByteBuffer(byteBuffer, MEDIAPIPE_IMAGE_FORMAT_RGBA, width, height)
    }

    fun rotateBitmap(bitmap: Bitmap, degrees: Int, mirror: Boolean = true): Bitmap {
        if (degrees == 0 && !mirror) return bitmap

        val matrix = Matrix()
        if (degrees != 0) {
            matrix.postRotate(degrees.toFloat())
        }
        if (mirror) {
            matrix.postScale(-1f, 1f)
        }

        return Bitmap.createBitmap(
            bitmap,
            0, 0,
            bitmap.width, bitmap.height,
            matrix,
            true
        )
    }

    fun resizeBitmap(bitmap: Bitmap, targetWidth: Int, targetHeight: Int): Bitmap {
        if (bitmap.width == targetWidth && bitmap.height == targetHeight) return bitmap

        return Bitmap.createScaledBitmap(bitmap, targetWidth, targetHeight, true)
    }

    fun centerCropBitmap(
        bitmap: Bitmap,
        targetWidth: Int,
        targetHeight: Int
    ): Bitmap {
        val sourceWidth = bitmap.width
        val sourceHeight = bitmap.height

        val sourceAspect = sourceWidth.toFloat() / sourceHeight
        val targetAspect = targetWidth.toFloat() / targetHeight

        val (cropX, cropY, cropW, cropH) = if (sourceAspect > targetAspect) {
            val croppedWidth = (sourceHeight * targetAspect).toInt()
            val x = (sourceWidth - croppedWidth) / 2
            intArrayOf(x, 0, croppedWidth, sourceHeight)
        } else {
            val croppedHeight = (sourceWidth / targetAspect).toInt()
            val y = (sourceHeight - croppedHeight) / 2
            intArrayOf(0, y, sourceWidth, croppedHeight)
        }

        val cropped = Bitmap.createBitmap(bitmap, cropX, cropY, cropW, cropH)
        return if (cropped.width == targetWidth && cropped.height == targetHeight) {
            cropped
        } else {
            resizeBitmap(cropped, targetWidth, targetHeight)
        }
    }
}
