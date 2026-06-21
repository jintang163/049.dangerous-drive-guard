package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dangerous-drive-guard/backend/internal/common/model"
	"github.com/dangerous-drive-guard/backend/pkg/database"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

type ImageEnhancementService struct {
	db *gorm.DB
}

func NewImageEnhancementService() *ImageEnhancementService {
	return &ImageEnhancementService{
		db: database.GetDB(),
	}
}

func (s *ImageEnhancementService) EnhanceImage(ctx context.Context, req *model.ImageEnhanceRequest) (*model.ImageEnhanceResult, error) {
	startTime := time.Now()

	img, _, err := decodeBase64Image(req.ImageBase64)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	cfg := s.getEffectiveConfig(req)

	origQuality := s.calcImageQuality(img)
	origBrightness := s.calcAvgBrightness(img)

	enhanced := img

	if cfg.HistogramEqApplied {
		enhanced = s.applyHistogramEqualization(enhanced)
	}

	if cfg.GammaValue != 1.0 && cfg.GammaValue > 0 {
		enhanced = s.applyGammaCorrection(enhanced, cfg.GammaValue)
	}

	if cfg.BrightnessDelta != 0 {
		enhanced = s.applyBrightness(enhanced, cfg.BrightnessDelta)
	}

	if cfg.ContrastDelta != 0 {
		enhanced = s.applyContrast(enhanced, cfg.ContrastDelta)
	}

	if cfg.DenoiseApplied {
		enhanced = s.applyDenoise(enhanced, 3)
	}

	if cfg.SharpenApplied {
		enhanced = s.applySharpen(enhanced)
	}

	enhQuality := s.calcImageQuality(enhanced)
	enhBrightness := s.calcAvgBrightness(enhanced)

	improvement := 0.0
	if origQuality > 0 {
		improvement = (enhQuality - origQuality) / origQuality * 100
	}

	base64Str, err := encodeImageToBase64(enhanced)
	if err != nil {
		return nil, fmt.Errorf("encode image: %w", err)
	}

	procMs := int(time.Since(startTime).Milliseconds())

	record := &model.ImageEnhancementRecord{
		VehicleID:           req.VehicleID,
		DriverID:            req.DriverID,
		WaybillID:           req.WaybillID,
		DeviceID:            req.DeviceID,
		OriginalImageURL:    req.ImageURL,
		EnhanceMode:         model.EnhanceMode(req.EnhanceMode),
		GammaValue:          &cfg.GammaValue,
		BrightnessDelta:     cfg.BrightnessDelta,
		ContrastDelta:       cfg.ContrastDelta,
		DenoiseApplied:      cfg.DenoiseApplied,
		DenoiseStrength:     3,
		HistogramEqApplied:  cfg.HistogramEqApplied,
		SharpenApplied:      cfg.SharpenApplied,
		OriginalBrightnessAvg: &origBrightness,
		EnhancedBrightnessAvg: &enhBrightness,
		LightLevelLux:       req.LightLevelLux,
		IsNightTime:         req.IsNightTime != nil && *req.IsNightTime,
		QualityScoreBefore:  origQuality,
		QualityScoreAfter:   enhQuality,
		QualityImprovementPct: improvement,
		ProcessingTimeMs:    procMs,
		ProcessOnEdge:       false,
		Timestamp:           time.Now(),
	}

	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			record.Timestamp = t
		}
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		logger.Global.Warn("save enhance record failed", zap.Error(err))
	}

	return &model.ImageEnhanceResult{
		EnhancedImageBase64: base64Str,
		EnhancedImageURL:    req.ImageURL,
		EnhanceMode:         req.EnhanceMode,
		GammaValue:          cfg.GammaValue,
		BrightnessDelta:     cfg.BrightnessDelta,
		ContrastDelta:       cfg.ContrastDelta,
		DenoiseApplied:      cfg.DenoiseApplied,
		HistogramEqApplied:  cfg.HistogramEqApplied,
		SharpenApplied:      cfg.SharpenApplied,
		QualityScoreBefore:  origQuality,
		QualityScoreAfter:   enhQuality,
		QualityImprovement:  improvement,
		OriginalBrightness:  origBrightness,
		EnhancedBrightness:  enhBrightness,
		ProcessingTimeMs:    procMs,
		RecordID:            record.ID,
	}, nil
}

type effectiveConfig struct {
	GammaValue         float64
	BrightnessDelta    int
	ContrastDelta      int
	DenoiseApplied     bool
	HistogramEqApplied bool
	SharpenApplied     bool
}

func (s *ImageEnhancementService) getEffectiveConfig(req *model.ImageEnhanceRequest) effectiveConfig {
	cfg := effectiveConfig{
		GammaValue:         1.2,
		BrightnessDelta:    30,
		ContrastDelta:      20,
		DenoiseApplied:     true,
		HistogramEqApplied: true,
		SharpenApplied:     false,
	}

	switch model.EnhanceMode(req.EnhanceMode) {
	case model.EnhanceModeNight:
		cfg.GammaValue = 1.2
		cfg.BrightnessDelta = 30
		cfg.ContrastDelta = 20
		cfg.DenoiseApplied = true
		cfg.HistogramEqApplied = true
	case model.EnhanceModeInfrared:
		cfg.GammaValue = 1.1
		cfg.BrightnessDelta = 15
		cfg.ContrastDelta = 25
		cfg.DenoiseApplied = true
		cfg.HistogramEqApplied = true
		cfg.SharpenApplied = true
	case model.EnhanceModeLowLight:
		cfg.GammaValue = 1.3
		cfg.BrightnessDelta = 40
		cfg.ContrastDelta = 25
		cfg.DenoiseApplied = true
		cfg.HistogramEqApplied = true
		cfg.SharpenApplied = true
	}

	if req.ApplyDenoise != nil {
		cfg.DenoiseApplied = *req.ApplyDenoise
	}
	if req.ApplySharpen != nil {
		cfg.SharpenApplied = *req.ApplySharpen
	}

	return cfg
}

func (s *ImageEnhancementService) calcAvgBrightness(img image.Image) int {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return 0
	}

	sampleStep := 4
	samples := 0
	totalBrightness := 0

	for y := 0; y < h; y += sampleStep {
		for x := 0; x < w; x += sampleStep {
			r, g, b, _ := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			lum := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			totalBrightness += int(lum)
			samples++
		}
	}

	if samples == 0 {
		return 0
	}
	return totalBrightness / samples
}

func (s *ImageEnhancementService) calcImageQuality(img image.Image) float64 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return 0
	}

	avgBrightness := float64(s.calcAvgBrightness(img))

	brightnessScore := 1.0
	if avgBrightness < 50 {
		brightnessScore = avgBrightness / 50.0
	} else if avgBrightness > 200 {
		brightnessScore = (255 - avgBrightness) / 55.0
	}

	contrastScore := 0.5
	sampleStep := 8
	edgeCount := 0
	totalPixels := 0

	for y := sampleStep; y < h; y += sampleStep {
		for x := sampleStep; x < w; x += sampleStep {
			px := img.At(x+bounds.Min.X, y+bounds.Min.Y)
			pxLeft := img.At(x+bounds.Min.X-sampleStep, y+bounds.Min.Y)
			pxUp := img.At(x+bounds.Min.X, y+bounds.Min.Y-sampleStep)

			r1, g1, b1, _ := px.RGBA()
			r2, g2, b2, _ := pxLeft.RGBA()
			r3, g3, b3, _ := pxUp.RGBA()

			lum1 := 0.299*float64(r1>>8) + 0.587*float64(g1>>8) + 0.114*float64(b1>>8)
			lum2 := 0.299*float64(r2>>8) + 0.587*float64(g2>>8) + 0.114*float64(b2>>8)
			lum3 := 0.299*float64(r3>>8) + 0.587*float64(g3>>8) + 0.114*float64(b3>>8)

			diff1 := math.Abs(lum1 - lum2)
			diff2 := math.Abs(lum1 - lum3)
			if diff1 > 20 || diff2 > 20 {
				edgeCount++
			}
			totalPixels++
		}
	}

	if totalPixels > 0 {
		edgeRatio := float64(edgeCount) / float64(totalPixels)
		contrastScore = math.Min(1.0, edgeRatio*3.0)
	}

	quality := 0.4*brightnessScore + 0.6*contrastScore
	return math.Max(0.05, math.Min(1.0, quality))
}

func (s *ImageEnhancementService) applyGammaCorrection(img image.Image, gamma float64) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	invGamma := 1.0 / gamma

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := uint8(math.Pow(float64(r>>8)/255.0, invGamma) * 255.0)
			ng := uint8(math.Pow(float64(g>>8)/255.0, invGamma) * 255.0)
			nb := uint8(math.Pow(float64(b>>8)/255.0, invGamma) * 255.0)
			out.SetRGBA(x, y, color.RGBA{nr, ng, nb, uint8(a >> 8)})
		}
	}
	return out
}

func (s *ImageEnhancementService) applyBrightness(img image.Image, delta int) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := clampInt(int(r>>8)+delta, 0, 255)
			ng := clampInt(int(g>>8)+delta, 0, 255)
			nb := clampInt(int(b>>8)+delta, 0, 255)
			out.SetRGBA(x, y, color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(a >> 8)})
		}
	}
	return out
}

func (s *ImageEnhancementService) applyContrast(img image.Image, delta int) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	factor := (259.0 * (float64(delta) + 255.0)) / (255.0 * (259.0 - float64(delta)))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := clampInt(int(factor*(float64(r>>8)-128)+128), 0, 255)
			ng := clampInt(int(factor*(float64(g>>8)-128)+128), 0, 255)
			nb := clampInt(int(factor*(float64(b>>8)-128)+128), 0, 255)
			out.SetRGBA(x, y, color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(a >> 8)})
		}
	}
	return out
}

func (s *ImageEnhancementService) applyHistogramEqualization(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	var histogram [256]int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			lum := int(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
			histogram[lum]++
		}
	}

	totalPixels := w * h
	var cdf [256]int
	cdf[0] = histogram[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + histogram[i]
	}

	var lut [256]uint8
	cdfMin := 0
	for i := 0; i < 256; i++ {
		if cdf[i] > 0 {
			cdfMin = cdf[i]
			break
		}
	}
	for i := 0; i < 256; i++ {
		if totalPixels-cdfMin > 0 {
			lut[i] = uint8(float64(cdf[i]-cdfMin) / float64(totalPixels-cdfMin) * 255.0)
		}
	}

	out := image.NewRGBA(bounds)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			lum := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			newLum := float64(lut[int(lum)])
			ratio := 1.0
			if lum > 0 {
				ratio = newLum / lum
			}
			nr := clampInt(int(float64(r>>8)*ratio), 0, 255)
			ng := clampInt(int(float64(g>>8)*ratio), 0, 255)
			nb := clampInt(int(float64(b>>8)*ratio), 0, 255)
			out.SetRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(a >> 8)})
		}
	}
	return out
}

func (s *ImageEnhancementService) applyDenoise(img image.Image, strength int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	out := image.NewRGBA(bounds)
	radius := strength

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sumR, sumG, sumB, count int
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx := x + dx
					ny := y + dy
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						r, g, b, _ := img.At(nx+bounds.Min.X, ny+bounds.Min.Y).RGBA()
						sumR += int(r >> 8)
						sumG += int(g >> 8)
						sumB += int(b >> 8)
						count++
					}
				}
			}
			if count > 0 {
				_, _, _, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
				out.SetRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.RGBA{
					uint8(sumR / count),
					uint8(sumG / count),
					uint8(sumB / count),
					uint8(a >> 8),
				})
			}
		}
	}
	return out
}

func (s *ImageEnhancementService) applySharpen(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	out := image.NewRGBA(bounds)

	kernel := [3][3]float64{
		{0, -1, 0},
		{-1, 5, -1},
		{0, -1, 0},
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sumR, sumG, sumB float64
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					nx := x + kx
					ny := y + ky
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						r, g, b, _ := img.At(nx+bounds.Min.X, ny+bounds.Min.Y).RGBA()
						k := kernel[ky+1][kx+1]
						sumR += float64(r>>8) * k
						sumG += float64(g>>8) * k
						sumB += float64(b>>8) * k
					}
				}
			}
			_, _, _, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			out.SetRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.RGBA{
				uint8(clampInt(int(sumR), 0, 255)),
				uint8(clampInt(int(sumG), 0, 255)),
				uint8(clampInt(int(sumB), 0, 255)),
				uint8(a >> 8),
			})
		}
	}
	return out
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func decodeBase64Image(base64Str string) (image.Image, string, error) {
	data := base64Str
	if strings.Contains(base64Str, ",") {
		parts := strings.SplitN(base64Str, ",", 2)
		data = parts[1]
	}

	imgBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", fmt.Errorf("base64 decode: %w", err)
	}

	reader := bytes.NewReader(imgBytes)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("image decode: %w", err)
	}

	return img, format, nil
}

func encodeImageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("jpeg encode: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *ImageEnhancementService) SaveEnhanceRecord(ctx context.Context, record *model.ImageEnhancementRecord) error {
	return s.db.WithContext(ctx).Create(record).Error
}

func (s *ImageEnhancementService) ListEnhanceRecords(ctx context.Context, vehicleID int64, page, pageSize int) ([]model.ImageEnhancementRecord, int64, error) {
	var records []model.ImageEnhancementRecord
	var total int64

	query := s.db.WithContext(ctx).Model(&model.ImageEnhancementRecord{})
	if vehicleID > 0 {
		query = query.Where("vehicle_id = ?", vehicleID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (s *ImageEnhancementService) GetEnhanceRecord(ctx context.Context, id int64) (*model.ImageEnhancementRecord, error) {
	var record model.ImageEnhancementRecord
	err := s.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
