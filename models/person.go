package models

// Person представляет информацию о человеке
// @Description Информация о человеке с обогащенными данными
type Person struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	CreatedAt  int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  int64  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  *int64 `json:"deleted_at,omitempty" gorm:"index"`
	Name       string `json:"name" gorm:"not null"`
	Surname    string `json:"surname" gorm:"not null"`
	Patronymic string `json:"patronymic,omitempty"`
	Age        int    `json:"age,omitempty"`
	Gender     string `json:"gender,omitempty"`
	Country    string `json:"country,omitempty"`
}

// PersonInput представляет входные данные для создания/обновления информации о человеке
// @Description Входные данные для создания или обновления информации о человеке
type PersonInput struct {
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Patronymic string `json:"patronymic,omitempty"`
}
