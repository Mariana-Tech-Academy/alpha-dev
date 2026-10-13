package service

import "alpha-dev/repository"

type AdminService interface {
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) AdminService {
	return &adminService{adminRepo: repo}
}
