package service

import "alpha-dev/repository"

type CategoryService interface {
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: repo}
}
