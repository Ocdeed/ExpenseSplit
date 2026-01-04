'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Receipt, Search } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';
import { Input } from '@/components/ui/input';

interface Expense {
  id: string;
  description: string;
  amount: number;
  paid_by: {
    name: string;
  };
  created_at: string;
  category: string;
  team_name: string;
  team_id: string;
}

export default function ExpensesPage() {
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    const fetchAllExpenses = async () => {
      try {
        const teamsRes = await api.get('/teams');
        const teams = teamsRes.data.data || [];
        
        const allExpenses: Expense[] = [];
        for (const team of teams) {
          const expRes = await api.get(`/teams/${team.id}/expenses`);
          const teamExpenses = (expRes.data.data || []).map((e: any) => ({
            ...e,
            team_name: team.name,
            team_id: team.id
          }));
          allExpenses.push(...teamExpenses);
        }
        
        // Sort by date descending
        allExpenses.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        setExpenses(allExpenses);
      } catch (error) {
        toast.error('Failed to fetch expenses');
      } finally {
        setIsLoading(false);
      }
    };

    fetchAllExpenses();
  }, []);

  const filteredExpenses = expenses.filter(e => 
    e.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
    e.team_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    e.category.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">All Expenses</h1>
        <div className="relative w-64">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search expenses..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Receipt className="w-5 h-5 mr-2" />
            Recent Expenses
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Team</TableHead>
                  <TableHead>Payer</TableHead>
                  <TableHead>Category</TableHead>
                  <TableHead className="text-right">Amount</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredExpenses.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-gray-500">No expenses found</TableCell>
                  </TableRow>
                ) : (
                  filteredExpenses.map((expense) => (
                    <TableRow key={expense.id}>
                      <TableCell>{new Date(expense.created_at).toLocaleDateString()}</TableCell>
                      <TableCell className="font-medium">{expense.description}</TableCell>
                      <TableCell>{expense.team_name}</TableCell>
                      <TableCell>{expense.paid_by.name}</TableCell>
                      <TableCell>{expense.category}</TableCell>
                      <TableCell className="text-right font-bold">${expense.amount.toFixed(2)}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
