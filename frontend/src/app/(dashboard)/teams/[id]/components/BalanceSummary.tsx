'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { BarChart3, ArrowRight, CheckCircle2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/badge';

interface BalanceSummaryProps {
  balances: any[];
  memberBalances: any[];
  onSettle: (from: string, to: string, amount: number) => void;
}

export function BalanceSummary({ balances, memberBalances, onSettle }: BalanceSummaryProps) {
  return (
    <div className="space-y-6">
      <Card className="glass-card border-none">
        <CardHeader>
          <CardTitle className="text-lg font-bold flex items-center gap-2">
            <BarChart3 className="w-5 h-5 text-primary" />
            Settlement Suggestions
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {balances.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <CheckCircle2 className="w-10 h-10 text-green-500/50 mb-2" />
              <p className="text-sm text-muted-foreground">All settled up! No pending payments.</p>
            </div>
          ) : (
            balances.map((b, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="p-4 rounded-2xl bg-primary/5 border border-primary/10 flex flex-col gap-3"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm">{b.from_user_name}</span>
                    <ArrowRight className="w-4 h-4 text-muted-foreground" />
                    <span className="font-semibold text-sm">{b.to_user_name}</span>
                  </div>
                  <span className="font-bold text-primary">${b.amount.toFixed(2)}</span>
                </div>
                <Button 
                  size="sm" 
                  className="w-full rounded-xl bg-primary hover:bg-primary/90"
                  onClick={() => onSettle(b.from_user, b.to_user, b.amount)}
                >
                  Mark as Settled
                </Button>
              </motion.div>
            ))
          )}
        </CardContent>
      </Card>

      <Card className="glass-card border-none">
        <CardHeader>
          <CardTitle className="text-lg font-bold">Member Balances</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {memberBalances.map((mb, idx) => (
            <div key={idx} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center text-xs font-bold">
                  {mb.name[0].toUpperCase()}
                </div>
                <span className="text-sm font-medium">{mb.name}</span>
              </div>
              <div className={`text-sm font-bold ${mb.net_balance >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                {mb.net_balance >= 0 ? '+' : ''}${mb.net_balance.toFixed(2)}
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
