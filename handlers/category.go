package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"alpha-dev/service"
	"alpha-dev/utils"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{CategoryService: categoryService}
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req createCategoryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid category JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid category JSON")
		return
	}

	category, err := h.CategoryService.CreateCategory(r.Context(), req.Name)
	if err != nil {
		var validationErr *service.CategoryValidationError
		switch {
		case errors.As(err, &validationErr):
			utils.ValidationError(w, "check the category name", validationErr.Fields)
		case errors.Is(err, service.ErrCategoryExists):
			utils.Error(w, http.StatusConflict, err.Error())
		default:
			log.Printf("create category: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to create category")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "category created", category)
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.CategoryService.GetCategories(r.Context())
	if err != nil {
		log.Printf("list categories: %v", err)
		utils.Error(w, http.StatusInternalServerError, "unable to load categories")
		return
	}
	utils.Success(w, http.StatusOK, "categories retrieved", categories)
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		utils.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	category, err := h.CategoryService.GetCategoryByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			utils.Error(w, http.StatusNotFound, "category not found")
			return
		}
		log.Printf("get category %d: %v", id, err)
		utils.Error(w, http.StatusInternalServerError, "unable to load category")
		return
	}

	utils.Success(w, http.StatusOK, "category retrieved", category)
}
