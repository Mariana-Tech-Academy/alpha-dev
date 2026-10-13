package handlers

import "alpha-dev/service"

type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{MenuService: service}

}
