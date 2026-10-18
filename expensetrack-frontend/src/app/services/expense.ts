
import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Expense {
  id?: number;
  title: string;
  category: string;
  amount: number;
  expense_date: string;
  description: string;
  payment_method: string;
}

export interface MonthlySummary {
  month: string;
  total_spent: number;
  expense_count: number;
  categories: { category: string; amount: number }[];
}

@Injectable({
  providedIn: 'root'
})
export class ExpenseService {
  private readonly apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  getExpenses(filters?: {
    category?: string;
    start_date?: string;
    end_date?: string;
  }): Observable<Expense[]> {
    let params = new HttpParams();

    if (filters?.category) {
      params = params.set('category', filters.category);
    }
    if (filters?.start_date) {
      params = params.set('start_date', filters.start_date);
    }
    if (filters?.end_date) {
      params = params.set('end_date', filters.end_date);
    }

    return this.http.get<Expense[]>(`${this.apiUrl}/expenses`, { params });
  }

  createExpense(expense: Expense): Observable<any> {
    return this.http.post(`${this.apiUrl}/expenses`, expense);
  }

  updateExpense(id: number, expense: Expense): Observable<any> {
    return this.http.put(`${this.apiUrl}/expenses/${id}`, expense);
  }

  deleteExpense(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/expenses/${id}`);
  }

  getSummary(month: string): Observable<MonthlySummary> {
    const params = new HttpParams().set('month', month);
    return this.http.get<MonthlySummary>(`${this.apiUrl}/summary`, { params });
  }
}
