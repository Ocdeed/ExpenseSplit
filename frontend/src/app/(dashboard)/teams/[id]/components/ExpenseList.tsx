'use client';

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Trash2, FileText, CheckCircle, XCircle, Clock } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { Card, CardContent } from '@/components/ui/card';

interface Expense {
  id: string;
  description: string;
  amount: number;
  paid_by: {
    id: string;
    name: string;
    email: string;
  };
  created_at: string;
  category: string;
  receipt_url?: string;
  approval_status: 'pending' | 'approved' | 'rejected';
}

interface ExpenseListProps {
  expenses: Expense[];
  onDelete: (id: string) => void;
}

const statusConfig = {
  pending: { icon: Clock, class: "bg-yellow-500/10 text-yellow-600 border-yellow-500/20" },
  approved: { icon: CheckCircle, class: "bg-green-500/10 text-green-600 border-green-500/20" },
  rejected: { icon: XCircle, class: "bg-red-500/10 text-red-600 border-red-500/20" },
};

export function ExpenseList({ expenses, onDelete }: ExpenseListProps) {
  return (
    <Card className="glass-card border-none overflow-hidden">
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader className="bg-secondary/30">
              <TableRow className="hover:bg-transparent border-border/50">
                <TableHead className="py-4">Date</TableHead>
                <TableHead className="py-4">Description</TableHead>
                <TableHead className="py-4">Payer</TableHead>
                <TableHead className="py-4">Category</TableHead>
                <TableHead className="py-4">Status</TableHead>
                <TableHead className="text-right py-4">Amount</TableHead>
                <TableHead className="text-right py-4">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <AnimatePresence mode="popLayout">
                {expenses.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-20 text-muted-foreground">
                      No expenses found for this team.
                    </TableCell>
                  </TableRow>
                ) : (
                  expenses.map((expense) => {
                    const StatusIcon = statusConfig[expense.approval_status].icon;
                    return (
                      <motion.tr
                        key={expense.id}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="group hover:bg-secondary/30 transition-colors border-border/50"
                      >
                        <TableCell className="py-4 text-muted-foreground text-sm">
                          {new Date(expense.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell className="py-4">
                          <div className="font-semibold">{expense.description}</div>
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
                          <Badge variant="secondary" className="capitalize font-normal">
                            {expense.category}
                          </Badge>
                        </TableCell>
                        <TableCell className="py-4">
                          <Badge className={statusConfig[expense.approval_status].class + " font-medium flex items-center w-fit gap-1"}>
                            <StatusIcon className="w-3 h-3" />
                            {expense.approval_status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right py-4 font-bold">
                          ${expense.amount.toFixed(2)}
                        </TableCell>
                        <TableCell className="text-right py-4">
                          <div className="flex justify-end gap-1">
                            {expense.receipt_url && (
                              <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
                                <a href={expense.receipt_url} target="_blank" rel="noopener noreferrer">
                                  <FileText className="w-4 h-4" />
                                </a>
                              </Button>
                            )}
                            <Button 
                              variant="ghost" 
                              size="icon" 
                              className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                              onClick={() => onDelete(expense.id)}
                            >
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </motion.tr>
                    );
                  })
                )}
              </AnimatePresence>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
