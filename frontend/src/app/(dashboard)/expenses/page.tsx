'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Receipt, Search, Filter, Download, Plus, Calendar } from 'lucide-react';
import api from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';

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

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.05 }
  }
};

const item = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0 }
};

export default function ExpensesPage() {
  const [searchTerm, setSearchTerm] = useState('');

  const { data: expenses = [], isLoading } = useQuery({
    queryKey: ['all-expenses'],
    queryFn: async () => {
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
      
      return allExpenses.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    }
  });

  const filteredExpenses = expenses.filter(e => 
    e.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
    e.team_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    e.category.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="space-y-8"
    >
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tight text-gradient">Expenses</h1>
          <p className="text-muted-foreground mt-1">Track and manage all your shared expenses across teams.</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" className="rounded-full">
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
          <Button className="rounded-full px-6 shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4 mr-2" />
            Add Expense
          </Button>
        </div>
      </div>

      <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
        <div className="relative w-full md:w-96 group">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
          <Input
            placeholder="Search by description, team, or category..."
            className="pl-10 bg-card/50 border-border/50 rounded-xl h-11 focus-visible:ring-primary/20"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-2 w-full md:w-auto">
          <Button variant="secondary" size="sm" className="rounded-lg h-11 px-4">
            <Filter className="w-4 h-4 mr-2" />
            Filters
          </Button>
          <Button variant="secondary" size="sm" className="rounded-lg h-11 px-4">
            <Calendar className="w-4 h-4 mr-2" />
            Date Range
          </Button>
        </div>
      </div>

      <Card className="glass-card border-none overflow-hidden">
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-6 space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-16 w-full rounded-xl" />
              ))}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader className="bg-secondary/30">
                  <TableRow className="hover:bg-transparent border-border/50">
                    <TableHead className="w-[120px] py-4">Date</TableHead>
                    <TableHead className="py-4">Description</TableHead>
                    <TableHead className="py-4">Team</TableHead>
                    <TableHead className="py-4">Payer</TableHead>
                    <TableHead className="py-4">Category</TableHead>
                    <TableHead className="text-right py-4">Amount</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <AnimatePresence mode="popLayout">
                    {filteredExpenses.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={6} className="text-center py-20">
                          <div className="flex flex-col items-center justify-center text-muted-foreground">
                            <Receipt className="w-12 h-12 mb-4 opacity-20" />
                            <p className="text-lg font-medium">No expenses found</p>
                            <p className="text-sm">Try adjusting your search or filters</p>
                          </div>
                        </TableCell>
                      </TableRow>
                    ) : (
                      filteredExpenses.map((expense) => (
                        <motion.tr
                          key={expense.id}
                          variants={item}
                          layout
                          className="group hover:bg-secondary/30 transition-colors border-border/50"
                        >
                          <TableCell className="py-4 text-muted-foreground">
                            {new Date(expense.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="font-semibold text-foreground group-hover:text-primary transition-colors">
                              {expense.description}
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <Badge variant="outline" className="bg-background/50 font-medium">
                              {expense.team_name}
                            </Badge>
                          </TableCell>
                          <TableCell className="py-4">
                            <div className="flex items-center gap-2">
                              <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center text-[10px] font-bold text-primary">
                                {expense.paid_by.name[0].toUpperCase()}
                              </div>
                              <span className="text-sm">{expense.paid_by.name}</span>
                            </div>
                          </TableCell>
                          <TableCell className="py-4">
                            <span className="text-sm text-muted-foreground capitalize">{expense.category}</span>
                          </TableCell>
                          <TableCell className="text-right py-4">
                            <span className="font-bold text-lg">${expense.amount.toFixed(2)}</span>
                          </TableCell>
                        </motion.tr>
                      ))
                    )}
                  </AnimatePresence>
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}
