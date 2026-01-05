'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { CheckCircle, XCircle, Clock, AlertCircle } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface Approval {
  id: string;
  expense_id: string;
  status: 'pending' | 'approved' | 'rejected';
  comment?: string;
  created_at: string;
  expense?: {
    description: string;
    amount: number;
    paid_by: { name: string };
  };
}

interface ApprovalListProps {
  approvals: Approval[];
  onUpdate: (id: string, status: 'approved' | 'rejected') => void;
}

export function ApprovalList({ approvals, onUpdate }: ApprovalListProps) {
  const pendingApprovals = approvals.filter(a => a.status === 'pending');

  if (pendingApprovals.length === 0) return null;

  return (
    <Card className="border-none bg-primary/5 shadow-none mb-8">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-bold flex items-center gap-2 text-primary">
          <AlertCircle className="w-4 h-4" />
          Pending Your Approval
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <AnimatePresence>
          {pendingApprovals.map((approval) => (
            <motion.div
              key={approval.id}
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="flex items-center justify-between p-3 rounded-xl bg-background border border-primary/10 shadow-sm"
            >
              <div className="flex flex-col">
                <span className="text-sm font-semibold">{approval.expense?.description}</span>
                <span className="text-xs text-muted-foreground">
                  ${approval.expense?.amount.toFixed(2)} paid by {approval.expense?.paid_by.name}
                </span>
              </div>
              <div className="flex gap-2">
                <Button 
                  size="sm" 
                  variant="ghost" 
                  className="h-8 w-8 p-0 text-red-500 hover:text-red-600 hover:bg-red-50"
                  onClick={() => onUpdate(approval.id, 'rejected')}
                >
                  <XCircle className="w-5 h-5" />
                </Button>
                <Button 
                  size="sm" 
                  variant="ghost" 
                  className="h-8 w-8 p-0 text-green-500 hover:text-green-600 hover:bg-green-50"
                  onClick={() => onUpdate(approval.id, 'approved')}
                >
                  <CheckCircle className="w-5 h-5" />
                </Button>
              </div>
            </motion.div>
          ))}
        </AnimatePresence>
      </CardContent>
    </Card>
  );
}
