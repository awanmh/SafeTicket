package http

import (
	"net/http"
	"safe-ticket/internal/domain"
	"safe-ticket/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	Usecase *usecase.EventUsecase
}

func NewEventHandler(r *gin.Engine, u *usecase.EventUsecase) {
	handler := &EventHandler{Usecase: u}
	
	r.POST("/book/unsafe", handler.BookUnsafe)
	r.POST("/book/safe", handler.BookSafe)
	r.GET("/events/:id", handler.GetEvent)
}

func (h *EventHandler) BookUnsafe(c *gin.Context) {
	// Kita samakan struct request-nya dengan yang Safe
	var req struct {
		EventID int    `json:"event_id"`
		UserID  string `json:"user_id"`
	}

	// 1. Binding JSON (Ini yang bikin error 400 tadi jika tidak ada)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// 2. Panggil Repo Unsafe
	// Di sini letak masalahnya: Code ini TIDAK pakai locking, jadi rawan balapan
	if err := h.Usecase.BookTicket(c.Request.Context(), domain.BookingRequest{EventID: req.EventID, UserID: req.UserID}, false); err != nil {
		// Jika sold out, kembalikan error (tapi di mode unsafe, ini sering telat)
		if err.Error() == "sold out" {
			c.JSON(http.StatusConflict, gin.H{"error": "Tiket habis"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking successful (Unsafe)"})
}

func (h *EventHandler) BookSafe(c *gin.Context) {
	var req domain.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if err := h.Usecase.BookTicket(c.Request.Context(), req, true); err != nil {
		// Jika errornya "no rows in result set", return 404 (Not Found)
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event tidak ditemukan"})
			return
		}

		// Jika errornya "sold out", return 409 (Conflict)
		if err.Error() == "sold out" {
			c.JSON(http.StatusConflict, gin.H{"error": "Tiket habis"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking successful (safe)"})
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	event, err := h.Usecase.GetEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}
