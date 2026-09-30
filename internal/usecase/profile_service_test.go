package usecase

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/store"
)

func TestProfileService_GetAndUpdate(t *testing.T) {
	saved := map[int64]*domain.UserProfile{1: {UserID: 1, ProfilePhoto: "http://x/old.png"}}
	users := &userRepoMock{
		getProfile:    func(id int64) (*domain.UserProfile, error) { return saved[id], nil },
		upsertProfile: func(p *domain.UserProfile) error { saved[p.UserID] = p; return nil },
	}
	svc := NewProfileService(&store.Store{UserRepo: users}, t.TempDir(), "http://api")

	if p, err := svc.GetProfile(1); err != nil || p.UserID != 1 {
		t.Errorf("GetProfile: %v", err)
	}

	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	updated, err := svc.UpdateProfile(1, "Ana", "Pop", &dob)
	if err != nil || updated.FirstName != "Ana" || updated.ProfilePhoto != "http://x/old.png" {
		t.Fatalf("UpdateProfile trebuie să păstreze poza: %v, %+v", err, updated)
	}
	if _, err := svc.UpdateProfile(2, "Ion", "", nil); err != nil || saved[2].ProfilePhoto != "" {
		t.Errorf("UpdateProfile pentru profil nou: %v", err)
	}

	users.upsertProfile = func(*domain.UserProfile) error { return errors.New("db down") }
	if _, err := svc.UpdateProfile(1, "x", "y", nil); err == nil {
		t.Error("eroarea de salvare trebuie propagată")
	}
}

func TestProfileService_UploadPhoto(t *testing.T) {
	saved := map[int64]*domain.UserProfile{}
	users := &userRepoMock{
		getProfile:    func(id int64) (*domain.UserProfile, error) { return saved[id], nil },
		upsertProfile: func(p *domain.UserProfile) error { saved[p.UserID] = p; return nil },
	}
	dir := filepath.Join(t.TempDir(), "avatars")
	svc := NewProfileService(&store.Store{UserRepo: users}, dir, "http://api")

	tmp, err := os.CreateTemp(t.TempDir(), "photo-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	if _, err := tmp.WriteString("png-bytes"); err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.UploadPhoto(1, tmp, &multipart.FileHeader{Filename: "a.gif", Size: 10}); err == nil {
		t.Error("extensia neacceptată trebuie să dea eroare")
	}
	if _, err := svc.UploadPhoto(1, tmp, &multipart.FileHeader{Filename: "a.png", Size: 6 * 1024 * 1024}); err == nil {
		t.Error("fișierul prea mare trebuie să dea eroare")
	}

	url, err := svc.UploadPhoto(1, tmp, &multipart.FileHeader{Filename: "a.png", Size: 9})
	if err != nil {
		t.Fatalf("UploadPhoto: %v", err)
	}
	if !strings.HasPrefix(url, "http://api/uploads/avatars/1_") || !strings.HasSuffix(url, ".png") {
		t.Errorf("URL greșit: %s", url)
	}
	if saved[1] == nil || saved[1].ProfilePhoto != url {
		t.Error("profilul trebuie actualizat cu noua poză")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("fișierul trebuie scris pe disc, am %d fișiere", len(entries))
	}

	// director imposibil de creat
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc = NewProfileService(&store.Store{UserRepo: users}, filepath.Join(blocked, "sub"), "http://api")
	if _, err := svc.UploadPhoto(1, tmp, &multipart.FileHeader{Filename: "a.png", Size: 9}); err == nil {
		t.Error("eroarea de creare a directorului trebuie propagată")
	}
}
