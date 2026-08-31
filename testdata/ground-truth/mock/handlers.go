package mock

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListCustomersHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "25")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		if limit <= 0 {
			limit = 25
		}
		data, total := store.ListCustomers(limit, offset)
		c.JSON(http.StatusOK, gin.H{
			"data":     data,
			"has_more": offset+limit < total,
			"total":    total,
		})
	}
}

func CreateCustomerHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Customer
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := store.CreateCustomer(req)
		c.JSON(http.StatusCreated, result)
	}
}

func GetCustomerHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		item, ok := store.GetCustomer(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func UpdateCustomerHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req Customer
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, ok := store.UpdateCustomer(id, req)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func ListInvoicesHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "25")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		if limit <= 0 {
			limit = 25
		}
		filterValue := c.Query("customer_id")
		data, total := store.ListInvoices(limit, offset, "customer_id", filterValue)
		c.JSON(http.StatusOK, gin.H{
			"data":     data,
			"has_more": offset+limit < total,
			"total":    total,
		})
	}
}

func CreateInvoiceHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Invoice
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := store.CreateInvoice(req)
		c.JSON(http.StatusCreated, result)
	}
}

func GetInvoiceHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		item, ok := store.GetInvoice(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func FinalizeInvoiceHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, ok := store.FinalizeInvoice(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func VoidInvoiceHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, ok := store.VoidInvoice(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func ListPaymentsHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "25")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		if limit <= 0 {
			limit = 25
		}
		filterValue := c.Query("invoice_id")
		data, total := store.ListPayments(limit, offset, "invoice_id", filterValue)
		c.JSON(http.StatusOK, gin.H{
			"data":     data,
			"has_more": offset+limit < total,
			"total":    total,
		})
	}
}

func CreatePaymentHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Payment
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := store.CreatePayment(req)
		c.JSON(http.StatusCreated, result)
	}
}

func GetPaymentHandler(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		item, ok := store.GetPayment(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// HealthHandler returns a simple health check endpoint.
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
