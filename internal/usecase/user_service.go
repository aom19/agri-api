package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
	"strings"
)

type UserService struct {
	users repository.UserRepository
	roles repository.RoleRepository
}

func NewUserService(users repository.UserRepository, roles repository.RoleRepository) *UserService {
	return &UserService{users: users, roles: roles}
}

func (s *UserService) GetAllUsers() ([]domain.User, error) {
	return s.users.GetAll()
}

func (s *UserService) GetUserByID(id int64) (*domain.User, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) GetUserDisplayName(id int64) (string, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		return "", err
	}
	if user == nil {
		users, err := s.users.GetAll()
		if err != nil {
			return "", err
		}
		for i := range users {
			if users[i].ID == id {
				user = &users[i]
				break
			}
		}
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	profile, err := s.users.GetProfile(id)
	if err == nil && profile != nil {
		name := strings.TrimSpace(strings.Join([]string{profile.FirstName, profile.LastName}, " "))
		if name != "" {
			return name, nil
		}
	}

	return user.Email, nil
}

func (s *UserService) CreateUser(email, password string, roleID int64, emailConfirmed bool) (*domain.User, error) {
	existing, err := s.users.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	role, err := s.roles.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	if password == "" {
		password = auth.GenerateRefreshToken()
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: hash,
		RoleID:       roleID,
		RoleCode:     role.Code,
		RoleName:     role.Name,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	if emailConfirmed {
		if err := s.users.MarkEmailConfirmed(user.ID); err != nil {
			return nil, err
		}
		user.EmailConfirmed = true
	}

	return s.GetUserByID(user.ID)
}

func (s *UserService) UpdateUser(id int64, email string, roleID int64, emailConfirmed bool, password string) (*domain.User, error) {
	current, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("user not found")
	}

	otherUser, err := s.users.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if otherUser != nil && otherUser.ID != id {
		return nil, errors.New("email already in use")
	}

	role, err := s.roles.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	current.Email = email
	current.RoleID = roleID
	current.RoleCode = role.Code
	current.RoleName = role.Name
	current.EmailConfirmed = emailConfirmed
	if err := s.users.Update(current); err != nil {
		return nil, err
	}

	if password != "" {
		hash, err := auth.HashPassword(password)
		if err != nil {
			return nil, err
		}
		if err := s.users.UpdatePassword(id, hash); err != nil {
			return nil, err
		}
	}

	return s.GetUserByID(id)
}

func (s *UserService) DeleteUser(id int64) error {
	user, err := s.users.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return s.users.Delete(id)
}

func (s *UserService) DisableUsers(id int64) error {
	user, err := s.users.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	return s.users.Disable(id)
}

func (s *UserService) EnableUsers(id int64) error {
	users, err := s.users.GetAll()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.ID != id {
			continue
		}
		if !user.Disabled {
			return errors.New("user already enabled")
		}
		return s.users.Enable(id)
	}

	return errors.New("user not found")
}
