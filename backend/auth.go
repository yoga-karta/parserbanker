package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const sessionTTL = 8 * time.Hour

// validUser/validPass diisi sekali oleh InitAuth() sebelum server listen.
var (
	validUser string
	validPass string
)

// InitAuth wajib dipanggil di main() sebelum app.Listen. Server MENOLAK START
// kalau CLIENT_USER/CLIENT_PASS belum di-set - tidak ada fallback hardcoded lagi
// (fallback lama, "Bni@2026#Xy9$Kz4!vL8pQw2R", sudah bocor karena ke-paste
// berkali-kali di chat - ganti password itu di server, bukan cuma di kode ini).
func InitAuth() {
	validUser = os.Getenv("CLIENT_USER")
	validPass = os.Getenv("CLIENT_PASS")
	if validUser == "" || validPass == "" {
		log.Fatal("CLIENT_USER dan CLIENT_PASS wajib di-set di environment - server tidak boleh jalan tanpa ini")
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sessionEntry struct {
	username string
	expires  time.Time
}

var (
	sessionsMu sync.Mutex
	sessions   = map[string]sessionEntry{}
)

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// handleLogin memvalidasi username/password terhadap CLIENT_USER/CLIENT_PASS,
// lalu menerbitkan token sesi acak (BUKAN password itu sendiri) yang kedaluwarsa
// otomatis setelah sessionTTL.
func handleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	userMatch := subtle.ConstantTimeCompare([]byte(req.Username), []byte(validUser)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(req.Password), []byte(validPass)) == 1
	if !userMatch || !passMatch {
		return c.Status(401).JSON(fiber.Map{"error": "Username atau Password salah!"})
	}

	token, err := newSessionToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal membuat sesi"})
	}

	sessionsMu.Lock()
	sessions[token] = sessionEntry{username: req.Username, expires: time.Now().Add(sessionTTL)}
	sessionsMu.Unlock()

	return c.JSON(fiber.Map{
		"token":      token,
		"username":   req.Username,
		"expires_in": int(sessionTTL.Seconds()),
	})
}

func handleLogout(c *fiber.Ctx) error {
	token := c.Get("X-API-Token")
	if token == "" {
		token = c.Query("token")
	}
	sessionsMu.Lock()
	delete(sessions, token)
	sessionsMu.Unlock()
	return c.JSON(fiber.Map{"ok": true})
}

// AuthMiddleware memvalidasi token sesi. Token ini beda dari password (acak,
// per-login, 32-byte) dan kedaluwarsa otomatis - jadi kalau nyangkut di access
// log nginx (lewat ?token= di link export), yang bocor cuma 1 sesi 8 jam,
// bukan kredensial permanen.
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("X-API-Token")
		if token == "" {
			token = c.Query("token")
		}

		sessionsMu.Lock()
		s, ok := sessions[token]
		if ok && time.Now().After(s.expires) {
			delete(sessions, token)
			ok = false
		}
		sessionsMu.Unlock()

		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: sesi login tidak valid atau kadaluarsa, silakan login ulang"})
		}

		c.Locals("username", s.username)
		return c.Next()
	}
}
