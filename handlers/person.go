package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"person-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgeResponse struct {
	Age int `json:"age"`
}

type GenderResponse struct {
	Gender string `json:"gender"`
}

type NationalityResponse struct {
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

func enrichPersonData(person *models.Person) error {
	// Get age
	resp, err := http.Get("https://api.agify.io/?name=" + person.Name)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ageResp AgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ageResp); err != nil {
		return err
	}
	person.Age = ageResp.Age

	// Get gender
	resp, err = http.Get("https://api.genderize.io/?name=" + person.Name)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var genderResp GenderResponse
	if err := json.NewDecoder(resp.Body).Decode(&genderResp); err != nil {
		return err
	}
	person.Gender = genderResp.Gender

	// Get nationality
	resp, err = http.Get("https://api.nationalize.io/?name=" + person.Name)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var nationalityResp NationalityResponse
	if err := json.NewDecoder(resp.Body).Decode(&nationalityResp); err != nil {
		return err
	}

	if len(nationalityResp.Country) > 0 {
		person.Country = nationalityResp.Country[0].CountryID
	}

	return nil
}

// CreatePerson godoc
// @Summary Создать нового человека
// @Description Создает нового человека и обогащает данные возрастом, полом и национальностью
// @Tags people
// @Accept json
// @Produce json
// @Param input body models.PersonInput true "Информация о человеке"
// @Success 201 {object} models.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people [post]
func CreatePerson(c *gin.Context, db *gorm.DB) {
	var input models.PersonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person := models.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
	}

	if err := enrichPersonData(&person); err != nil {
		log.Printf("Error enriching person data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enrich person data"})
		return
	}

	if err := db.Create(&person).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, person)
}

// GetPeople godoc
// @Summary Получить список людей
// @Description Получает список людей с возможностью фильтрации и пагинации
// @Tags people
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество записей на странице" default(10)
// @Param name query string false "Фильтр по имени"
// @Param surname query string false "Фильтр по фамилии"
// @Param gender query string false "Фильтр по полу"
// @Param country query string false "Фильтр по стране"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /people [get]
func GetPeople(c *gin.Context, db *gorm.DB) {
	var people []models.Person
	query := db.Model(&models.Person{})

	// Add filters
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if surname := c.Query("surname"); surname != "" {
		query = query.Where("surname ILIKE ?", "%"+surname+"%")
	}
	if gender := c.Query("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}
	if country := c.Query("country"); country != "" {
		query = query.Where("country = ?", country)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&people).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": people,
		"meta": gin.H{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// UpdatePerson godoc
// @Summary Обновить информацию о человеке
// @Description Обновляет информацию о человеке по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param input body models.PersonInput true "Обновленная информация"
// @Success 200 {object} models.Person
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [put]
func UpdatePerson(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var person models.Person

	if err := db.First(&person, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		return
	}

	var input models.PersonInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person.Name = input.Name
	person.Surname = input.Surname
	person.Patronymic = input.Patronymic

	if err := enrichPersonData(&person); err != nil {
		log.Printf("Error enriching person data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enrich person data"})
		return
	}

	if err := db.Save(&person).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, person)
}

// DeletePerson godoc
// @Summary Удалить человека
// @Description Удаляет человека по ID
// @Tags people
// @Produce json
// @Param id path int true "ID человека"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [delete]
func DeletePerson(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	if err := db.Delete(&models.Person{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Person deleted successfully"})
}
