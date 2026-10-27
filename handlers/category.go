package handlers

import (
	"alpha-dev/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{CategoryService: service}

} // Create the category, Decode JSON, return 201
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	cats, err := h.CategoryService.CreateCategory(input.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "app,lication/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cats)
}

// Get all Category and return them as JSOn
func (h *CategoryHandler) GetGategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.CategoryService.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cats)
}

// Get the ID from the URL , fetch that particular category and return it as JSON
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cat, err := h.CategoryService.GetCategoryByID(uint(id))
	if err != nil {
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)

}
