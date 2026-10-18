
# Expense Tracker

A full-stack Expense Tracker application built using Go, Gin, MySQL, and Angular.

## Features
- Add, view, edit, and delete expenses
- Filter expenses by category and date
- Monthly spending summary
- REST API backend
- Angular frontend

## Tech Stack
- Frontend: Angular, TypeScript, HTML, CSS
- Backend: Go, Gin
- Database: MySQL

## Project Structure
- backend/ — Go REST API
- frontend/ — Angular application

## Setup
1. Create a MySQL database named `expense_tracker`.
2. Create the expenses table using the SQL schema provided with the project.
3. Configure the `DB_DSN` environment variable.
4. Start the Go backend from the backend directory.
5. Install frontend dependencies using `npm install` inside frontend.
6. Start Angular using `ng serve`.

Backend: http://localhost:8080  
Frontend: http://localhost:4200

## API Endpoints
- GET /api/health
- GET /api/expenses
- POST /api/expenses
- PUT /api/expenses/:id
- DELETE /api/expenses/:id
- GET /api/summary?month=YYYY-MM
