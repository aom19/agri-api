package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type ProfileService struct {
	users     repository.UserRepository
	uploadDir string
	baseURL   string
}

func NewProfileService(users repository.UserRepository, uploadDir, baseURL string) *ProfileService {
	return &ProfileService{users: users, uploadDir: uploadDir, baseURL: baseURL}
}

func (svc *ProfileService) GetProfile(userID int64) (*domain.UserProfile, error) {
	return svc.users.GetProfile(userID)
}

func (svc *ProfileService) UpdateProfile(userID int64, firstName, lastName string, dob *time.Time) (*domain.UserProfile, error) {
	profile := &domain.UserProfile{
		UserID:      userID,
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: dob,
	}

	// Preserve existing photo
	existing, err := svc.users.GetProfile(userID)
	if err == nil && existing != nil {
		profile.ProfilePhoto = existing.ProfilePhoto
	}

	if err := svc.users.UpsertProfile(profile); err != nil {
		return nil, err
	}
	return svc.users.GetProfile(userID)
}

func (svc *ProfileService) UploadPhoto(userID int64, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		return "", fmt.Errorf("tip de fișier neacceptat: %s", ext)
	}
	if header.Size > 5*1024*1024 {
		return "", fmt.Errorf("fișierul depășește limita de 5MB")
	}

	if err := os.MkdirAll(svc.uploadDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	dst := filepath.Join(svc.uploadDir, filename)

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	photoURL := fmt.Sprintf("%s/uploads/avatars/%s", svc.baseURL, filename)

	// Update only the photo field
	profile, _ := svc.users.GetProfile(userID)
	if profile == nil {
		profile = &domain.UserProfile{UserID: userID}
	}
	profile.ProfilePhoto = photoURL
	_ = svc.users.UpsertProfile(profile)

	return photoURL, nil
}
