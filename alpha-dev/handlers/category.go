package handlers

import "alpha-dev/service"

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{CategoryService: service}

}
