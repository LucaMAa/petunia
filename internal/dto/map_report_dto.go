package dto

import "petunia/internal/model"

type CreateReportDto struct {
	Type        model.ReportType `json:"type"        binding:"required"`
	Title       string           `json:"title"       binding:"required"`
	Description string           `json:"description"`
	Lat         float64          `json:"lat"         binding:"required"`
	Lng         float64          `json:"lng"         binding:"required"`
	ImageURLs   []string         `json:"image_urls"`
}

type NearbyReportsParams struct {
	Lat    float64 `form:"lat"    binding:"required"`
	Lng    float64 `form:"lng"    binding:"required"`
	Radius float64 `form:"radius"`
}

type ReportAbuseDto struct {
	Reason string `json:"reason"`
}

type ReportResponse struct {
	*model.MapReport
	ImageURLs []string `json:"image_urls"`
	Distance  float64  `json:"distance_m,omitempty"`
}
