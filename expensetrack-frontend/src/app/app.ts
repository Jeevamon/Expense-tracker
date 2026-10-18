
import { CommonModule } from '@angular/common';
import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import {
  Expense,
  ExpenseService,
  MonthlySummary
} from './services/expense';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App implements OnInit {
  expenses: Expense[] = [];
  summary: MonthlySummary | null = null;

  categories = ['Food', 'Travel', 'Shopping', 'Utilities', 'Other'];
  paymentMethods = ['Cash', 'UPI', 'Card', 'Bank Transfer', 'Other'];

  selectedCategory = '';
  startDate = '';
  endDate = '';
  selectedMonth = new Date().toISOString().slice(0, 7);

  loading = false;
  saving = false;
  errorMessage = '';
  successMessage = '';
  editingId: number | null = null;

  form: Expense = this.emptyForm();

  constructor(
    private expenseService: ExpenseService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    // Load the dashboard automatically when the app opens.
    this.loadExpenses();
    this.loadSummary();
  }

  private refreshView(): void {
    // Required to notify Angular in zoneless applications.
    this.cdr.markForCheck();
  }

  private emptyForm(): Expense {
    return {
      title: '',
      category: 'Food',
      amount: 0,
      expense_date: new Date().toISOString().slice(0, 10),
      description: '',
      payment_method: 'UPI'
    };
  }

  // Load expenses using the current filters.
  loadExpenses(): void {
    this.loading = true;
    this.errorMessage = '';
    this.refreshView();

    this.expenseService.getExpenses({
      category: this.selectedCategory || undefined,
      start_date: this.startDate || undefined,
      end_date: this.endDate || undefined
    }).subscribe({
      next: (data) => {
        this.expenses = data ?? [];
        this.loading = false;
        this.refreshView();
      },
      error: (error) => {
        console.error('Failed to load expenses:', error);
        this.expenses = [];
        this.loading = false;
        this.errorMessage = 'Failed to load expenses. Please try again.';
        this.refreshView();
      }
    });
  }

  // Load the summary for the selected month.
  loadSummary(): void {
    if (!this.selectedMonth) {
      this.summary = null;
      this.refreshView();
      return;
    }

    this.expenseService.getSummary(this.selectedMonth).subscribe({
      next: (data) => {
        this.summary = data;
        this.refreshView();
      },
      error: (error) => {
        console.error('Failed to load monthly summary:', error);
        this.summary = null;
        this.refreshView();
      }
    });
  }

  // Apply category and date filters.
  applyFilters(): void {
    this.loadExpenses();
  }

  // Clear filters and reload all expenses.
  clearFilters(): void {
    this.selectedCategory = '';
    this.startDate = '';
    this.endDate = '';
    this.loadExpenses();
  }

  // Reload the summary only when the selected month changes.
  changeMonth(): void {
    this.loadSummary();
  }

  // Create a new expense or update an existing one.
  saveExpense(): void {
    this.errorMessage = '';
    this.successMessage = '';

    if (
      !this.form.title.trim() ||
      !this.form.expense_date ||
      !this.form.category ||
      !this.form.payment_method ||
      !Number.isFinite(Number(this.form.amount)) ||
      Number(this.form.amount) <= 0
    ) {
      this.errorMessage =
        'Enter a title, valid date, category, payment method and amount greater than zero.';
      this.refreshView();
      return;
    }

    const payload: Expense = {
      ...this.form,
      title: this.form.title.trim(),
      amount: Number(this.form.amount),
      description: this.form.description ?? ''
    };

    this.saving = true;
    this.refreshView();

    const isEditing = this.editingId !== null;

    const request = isEditing
      ? this.expenseService.updateExpense(this.editingId!, payload)
      : this.expenseService.createExpense(payload);

    request.subscribe({
      next: () => {
        this.saving = false;
        this.successMessage = isEditing
          ? 'Expense updated successfully.'
          : 'Expense added successfully.';

        this.cancelEdit();
        this.loadExpenses();
        this.loadSummary();
        this.refreshView();
      },
      error: (error) => {
        console.error('Failed to save expense:', error);
        this.saving = false;
        this.errorMessage =
          error.error?.error ||
          'Could not save the expense. Please check the backend API.';
        this.refreshView();
      }
    });
  }

  // Populate the form for editing.
  editExpense(expense: Expense): void {
    if (expense.id === undefined) {
      return;
    }

    this.editingId = expense.id;

    this.form = {
      ...expense,
      expense_date: String(expense.expense_date).slice(0, 10),
      description: expense.description ?? ''
    };

    this.successMessage = '';
    this.errorMessage = '';
    this.refreshView();

    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  }

  // Reset the form and exit edit mode.
  cancelEdit(): void {
    this.editingId = null;
    this.form = this.emptyForm();
    this.refreshView();
  }

  // Delete an expense after confirmation.
  deleteExpense(expense: Expense): void {
    if (expense.id === undefined) {
      return;
    }

    if (!window.confirm(`Delete "${expense.title}"?`)) {
      return;
    }

    this.errorMessage = '';
    this.successMessage = '';

    this.expenseService.deleteExpense(expense.id).subscribe({
      next: () => {
        this.successMessage = 'Expense deleted successfully.';
        this.loadExpenses();
        this.loadSummary();
        this.refreshView();
      },
      error: (error) => {
        console.error('Failed to delete expense:', error);
        this.errorMessage = 'Could not delete this expense.';
        this.refreshView();
      }
    });
  }

  // Calculate the total of currently displayed expenses.
  get totalVisibleExpenses(): number {
    return this.expenses.reduce(
      (total, expense) => total + Number(expense.amount),
      0
    );
  }

  // Track rows efficiently in the expenses table.
  trackById(index: number, expense: Expense): number {
    return expense.id ?? index;
  }
}


