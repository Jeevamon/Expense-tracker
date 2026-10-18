
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Read database connection settings
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Println("MySQL database connected successfully")

	// Create Gin router
	router := gin.Default()

	// CORS middleware: allow the Angular frontend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"message":  "Expense Tracker API is running",
			"database": "connected",
		})
	})

	// Create expense
	router.POST("/api/expenses", func(c *gin.Context) {
		var expense struct {
			Title         string  `json:"title" binding:"required"`
			Category      string  `json:"category" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,gt=0"`
			ExpenseDate   string  `json:"expense_date" binding:"required"`
			Description   string  `json:"description"`
			PaymentMethod string  `json:"payment_method" binding:"required"`
		}

		if err := c.ShouldBindJSON(&expense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid expense data. Check required fields and amount.",
			})
			return
		}

		result, err := db.Exec(`
			INSERT INTO expenses
			(title, category, amount, expense_date, description, payment_method)
			VALUES (?, ?, ?, ?, ?, ?)`,
			expense.Title,
			expense.Category,
			expense.Amount,
			expense.ExpenseDate,
			expense.Description,
			expense.PaymentMethod,
		)

		if err != nil {
			log.Println("Insert expense error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save expense",
			})
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println("Get inserted ID error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Expense saved, but failed to retrieve its ID",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Expense created successfully",
			"id":      id,
		})
	})

	// Get expenses with optional filters
	router.GET("/api/expenses", func(c *gin.Context) {
		query := `
			SELECT id, title, category, amount, expense_date,
			       description, payment_method
			FROM expenses
			WHERE 1=1
		`

		args := []interface{}{}

		category := c.Query("category")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		if category != "" {
			query += " AND category = ?"
			args = append(args, category)
		}

		if startDate != "" {
			query += " AND expense_date >= ?"
			args = append(args, startDate)
		}

		if endDate != "" {
			query += " AND expense_date <= ?"
			args = append(args, endDate)
		}

		query += " ORDER BY expense_date DESC, id DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Println("Query error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve expenses",
			})
			return
		}
		defer rows.Close()

		expenses := []gin.H{}

		for rows.Next() {
			var id int
			var title, category, description, paymentMethod string
			var amount float64
			var expenseDate time.Time

			if err := rows.Scan(
				&id,
				&title,
				&category,
				&amount,
				&expenseDate,
				&description,
				&paymentMethod,
			); err != nil {
				log.Println("Scan expense error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to read expense",
				})
				return
			}

			expenses = append(expenses, gin.H{
				"id":             id,
				"title":          title,
				"category":       category,
				"amount":         amount,
				"expense_date":   expenseDate.Format("2006-01-02"),
				"description":    description,
				"payment_method": paymentMethod,
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("Rows error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read expenses",
			})
			return
		}

		c.JSON(http.StatusOK, expenses)
	})

	// Update expense
	router.PUT("/api/expenses/:id", func(c *gin.Context) {
		id := c.Param("id")

		if parsedID, err := strconv.Atoi(id); err != nil || parsedID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Expense ID must be a positive integer",
			})
			return
		}

		var expense struct {
			Title         string  `json:"title" binding:"required"`
			Category      string  `json:"category" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,gt=0"`
			ExpenseDate   string  `json:"expense_date" binding:"required"`
			Description   string  `json:"description"`
			PaymentMethod string  `json:"payment_method" binding:"required"`
		}

		if err := c.ShouldBindJSON(&expense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid expense data",
			})
			return
		}

		result, err := db.Exec(`
			UPDATE expenses
			SET title = ?, category = ?, amount = ?,
			    expense_date = ?, description = ?, payment_method = ?
			WHERE id = ?`,
			expense.Title,
			expense.Category,
			expense.Amount,
			expense.ExpenseDate,
			expense.Description,
			expense.PaymentMethod,
			id,
		)

		if err != nil {
			log.Println("Update error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update expense",
			})
			return
		}

		affected, err := result.RowsAffected()
		if err != nil {
			log.Println("Rows affected error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not verify update",
			})
			return
		}

		if affected == 0 {
			var exists bool
			err := db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM expenses WHERE id = ?)", id,
			).Scan(&exists)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Could not verify expense",
				})
				return
			}

			if !exists {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Expense not found",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Expense updated successfully",
		})
	})

	// Delete expense
	router.DELETE("/api/expenses/:id", func(c *gin.Context) {
		id := c.Param("id")

		if parsedID, err := strconv.Atoi(id); err != nil || parsedID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Expense ID must be a positive integer",
			})
			return
		}

		result, err := db.Exec(
			"DELETE FROM expenses WHERE id = ?",
			id,
		)

		if err != nil {
			log.Println("Delete error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete expense",
			})
			return
		}

		affected, err := result.RowsAffected()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not verify deletion",
			})
			return
		}

		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Expense not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Expense deleted successfully",
		})
	})

	// Monthly summary and category totals
	router.GET("/api/summary", func(c *gin.Context) {
		month := c.DefaultQuery("month", time.Now().Format("2006-01"))

		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil || parsedMonth.Format("2006-01") != month {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Month must use YYYY-MM format",
			})
			return
		}

		startDate := parsedMonth.Format("2006-01") + "-01"
		endDate := parsedMonth.AddDate(0, 1, 0).Format("2006-01") + "-01"

		var total float64
		var count int

		err = db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0), COUNT(*)
			FROM expenses
			WHERE expense_date >= ? AND expense_date < ?
		`, startDate, endDate).Scan(&total, &count)

		if err != nil {
			log.Println("Summary error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to calculate summary",
			})
			return
		}

		rows, err := db.Query(`
			SELECT category, SUM(amount)
			FROM expenses
			WHERE expense_date >= ? AND expense_date < ?
			GROUP BY category
			ORDER BY SUM(amount) DESC
		`, startDate, endDate)

		if err != nil {
			log.Println("Category summary error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve category summary",
			})
			return
		}
		defer rows.Close()

		categories := []gin.H{}

		for rows.Next() {
			var category string
			var amount float64

			if err := rows.Scan(&category, &amount); err != nil {
				log.Println("Category scan error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to read category summary",
				})
				return
			}

			categories = append(categories, gin.H{
				"category": category,
				"amount":   amount,
			})
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read category summary",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"month":         month,
			"total_spent":   total,
			"expense_count": count,
			"categories":    categories,
		})
	})

	// Start server
	log.Println("Registered API routes:")
	log.Println("GET    /api/health")
	log.Println("GET    /api/expenses")
	log.Println("POST   /api/expenses")
	log.Println("PUT    /api/expenses/:id")
	log.Println("DELETE /api/expenses/:id")
	log.Println("GET    /api/summary")
	log.Println("Server running at http://localhost:8080")

	log.Fatal(router.Run(":8080"))
}
