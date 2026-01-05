'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3, TrendingUp, TrendingDown, Wallet, ArrowRight } from 'lucide-react';
import api from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Skeleton } from '@/components/ui/skeleton';
import Link from 'next/link';

interface TeamBalance {
  team_id: string;
  team_name: string;
  balance: number;
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.1 }
  }
};

const item = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 }
};

export default function BalancesPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['all-balances'],
    queryFn: async () => {
      const teamsRes = await api.get('/teams');
      const teams = teamsRes.data.data || [];
      
      const balances: TeamBalance[] = [];
      let owe = 0;
      let owed = 0;

      for (const team of teams) {
        const balRes = await api.get(`/teams/${team.id}/balances/me`);
        const amount = balRes.data.data.net_balance;
        balances.push({
          team_id: team.id,
          team_name: team.name,
          balance: amount
        });
        
        if (amount < 0) owe += Math.abs(amount);
        else owed += amount;
      }

      return { balances, totalOwe: owe, totalOwed: owed };
    }
  });

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="space-y-8"
    >
      <div>
        <h1 className="text-4xl font-bold tracking-tight text-gradient">Your Balances</h1>
        <p className="text-muted-foreground mt-1">A detailed breakdown of what you owe and what's owed to you.</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <motion.div variants={item}>
          <Card className="border-none bg-red-500/5 overflow-hidden group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-red-600">Total You Owe</CardTitle>
              <div className="p-2 bg-red-500/10 rounded-lg group-hover:bg-red-500/20 transition-colors">
                <TrendingDown className="h-4 w-4 text-red-600" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? <Skeleton className="h-8 w-24" /> : (
                <div className="text-3xl font-bold text-red-600">${data?.totalOwe.toFixed(2)}</div>
              )}
              <p className="text-xs text-red-600/60 mt-1">Across all active teams</p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div variants={item}>
          <Card className="border-none bg-green-500/5 overflow-hidden group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-green-600">Total You Are Owed</CardTitle>
              <div className="p-2 bg-green-500/10 rounded-lg group-hover:bg-green-500/20 transition-colors">
                <TrendingUp className="h-4 w-4 text-green-600" />
              </div>
            </CardHeader>
            <CardContent>
              {isLoading ? <Skeleton className="h-8 w-24" /> : (
                <div className="text-3xl font-bold text-green-600">${data?.totalOwed.toFixed(2)}</div>
              )}
              <p className="text-xs text-green-600/60 mt-1">To be collected from members</p>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      <motion.div variants={item}>
        <Card className="glass-card border-none overflow-hidden">
          <CardHeader className="border-b border-border/50 bg-secondary/30">
            <CardTitle className="text-lg font-bold flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-primary" />
              Balances by Team
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-16 w-full rounded-xl" />
                ))}
              </div>
            ) : (
              <div className="divide-y divide-border/50">
                {data?.balances.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
                    <Wallet className="w-12 h-12 mb-4 opacity-20" />
                    <p className="text-lg font-medium">No balances found</p>
                    <p className="text-sm">Join a team to start tracking balances.</p>
                  </div>
                ) : (
                  data?.balances.map((tb) => (
                    <Link key={tb.team_id} href={`/teams/${tb.team_id}`}>
                      <div className="flex items-center justify-between p-6 hover:bg-secondary/30 transition-all group">
                        <div className="flex items-center gap-4">
                          <div className="w-10 h-10 rounded-xl bg-primary/5 flex items-center justify-center group-hover:bg-primary/10 transition-colors">
                            <BarChart3 className="w-5 h-5 text-primary" />
                          </div>
                          <div>
                            <div className="font-bold text-foreground group-hover:text-primary transition-colors">{tb.team_name}</div>
                            <div className="text-xs text-muted-foreground">Click to view team details</div>
                          </div>
                        </div>
                        <div className="flex items-center gap-6">
                          <div className={`text-lg font-bold ${tb.balance >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                            {tb.balance >= 0 ? '+' : ''}${tb.balance.toFixed(2)}
                          </div>
                          <ArrowRight className="w-4 h-4 text-muted-foreground opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all" />
                        </div>
                      </div>
                    </Link>
                  ))
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </motion.div>
  );
}
