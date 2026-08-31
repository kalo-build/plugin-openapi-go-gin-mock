package mock

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all generated mock handlers on the Gin engine.
func RegisterRoutes(r *gin.Engine, store *Store) {
	r.GET("/health", HealthHandler())
	r.GET("/v1/customers", ListCustomersHandler(store))
	r.POST("/v1/customers", CreateCustomerHandler(store))
	r.GET("/v1/customers/:id", GetCustomerHandler(store))
	r.PUT("/v1/customers/:id", UpdateCustomerHandler(store))
	r.GET("/v1/invoices", ListInvoicesHandler(store))
	r.POST("/v1/invoices", CreateInvoiceHandler(store))
	r.GET("/v1/invoices/:id", GetInvoiceHandler(store))
	r.POST("/v1/invoices/:id/finalize", FinalizeInvoiceHandler(store))
	r.POST("/v1/invoices/:id/void", VoidInvoiceHandler(store))
	r.GET("/v1/payments", ListPaymentsHandler(store))
	r.POST("/v1/payments", CreatePaymentHandler(store))
	r.GET("/v1/payments/:id", GetPaymentHandler(store))
}
