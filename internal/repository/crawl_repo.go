package repository

import (
	"github.com/MrBista/The-Crawler/internal/models"
	"gorm.io/gorm"
)

type CrawlRepository interface {
	SavePage(page *models.CrawlPage) error
}

type CrawlRepositoryImpl struct {
	DB *gorm.DB
}

func NewCrawlRepositoryImpl(db *gorm.DB) *CrawlRepositoryImpl {
	return &CrawlRepositoryImpl{
		DB: db,
	}
}

func (r *CrawlRepositoryImpl) SavePage(page *models.CrawlPage) error {
	return r.DB.Save(page).Error
}
