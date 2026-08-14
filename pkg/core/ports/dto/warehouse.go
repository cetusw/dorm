package dto

type WarehouseItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
}

type WarehouseResponse struct {
	Items []WarehouseItem `json:"items"`
}

type WarehouseMovementActor struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type WarehouseMovement struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Quantity     int64                  `json:"quantity"`
	Comment      *string                `json:"comment"`
	CreatedBy    WarehouseMovementActor `json:"created_by"`
	CreatedAt    string                 `json:"created_at"`
	BalanceAfter int64                  `json:"balance_after"`
}

type WarehouseItemHistoryResponse struct {
	Item      WarehouseItem       `json:"item"`
	Movements []WarehouseMovement `json:"movements"`
}

type CreateWarehouseItemRequest struct {
	Name     string  `json:"name"`
	Quantity *int64  `json:"quantity"`
	Comment  *string `json:"comment"`
}

type UpdateWarehouseItemRequest struct {
	Name string `json:"name"`
}

type WarehouseMovementRequest struct {
	Quantity int64   `json:"quantity"`
	Comment  *string `json:"comment"`
}
