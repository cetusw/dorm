package dto

import "github.com/google/uuid"

type TaskCatalogItem struct {
	ID        uuid.UUID
	AreaID    int
	AreaName  string
	Title     string
	Cost      int
	Frequency int
}

type UpsertTaskCatalogRequest struct {
	AreaID    int    `form:"area_id"`
	Title     string `form:"title"`
	Cost      int    `form:"cost"`
	Frequency int    `form:"frequency"`
}

