package service

import (
	"context"
	v1 "rest_demo/api/v1"
	"rest_demo/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type LoginService interface {
	Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error)
}

func NewLoginService(service *Service, userRepo repository.SysUserRepository) LoginService {
	return &loginService{
		Service:  service,
		userRepo: userRepo,
	}
}

type loginService struct {
	userRepo repository.SysUserRepository
	*Service
}

func (s *loginService) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error) {

	m, err := s.userRepo.GetByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil || m == nil {
		return nil, nil
	}

	if bcrypt.CompareHashAndPassword(m.Password, []byte(req.Password)) != nil {
	}

	return nil, nil
}
