'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Receipt, Users, TrendingUp, TrendingDown, Plus, ArrowUpRight, ArrowDownRight, Wallet } from 'lucide-react';
import api from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Skeleton } from '@/components/ui/skeleton';

interface DashboardStats {
  totalExpenses: number;
  totalTeams: number;
  youOwe: number;
  youAreOwed: number;
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1
    }
  }
};

const item = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 }
};

export default function DashboardPage() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const teamsRes = await api.get('/teams');
      const teams = teamsRes.data.data || [];
      
      let totalOwe = 0;
      let totalOwed = 0;
      
      for (const team of teams) {
        const balanceRes = await api.get(`/teams/${team.id}/balances/me`);
        const balance = balanceRes.data.data;
        if (balance.net_balance < 0) totalOwe += Math.abs(balance.net_balance);
        else totalOwed += balance.net_balance;
      }

      return {
        totalExpenses: 0,
        totalTeams: teams.length,
        youOwe: totalOwe,
        youAreOwed: totalOwed,
      } as DashboardStats;
    }
  });

  if (isLoading) {
    return (
      <div className="space-y-8">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32 w-full rounded-2xl" />
          ))}
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
          <Skeleton className="col-span-4 h-[400px] rounded-2xl" />
          <Skeleton className="col-span-3 h-[400px] rounded-2xl" />
        </div>
      </div>
    );
  }

  const netBalance = (stats?.youAreOwed || 0) - (stats?.youOwe || 0);

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="space-y-8"
    >
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tight text-gradient">Dashboard</h1>
          <p className="text-muted-foreground mt-1">Welcome back! Here's what's happening with your expenses.</p>
        </div>
        <div className="flex items-center gap-3">
          <Button className="rounded-full px-6 shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4 mr-2" />
            Add Expense
          </Button>
        </div>
      </div>
      
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <motion.div variants={item}>
          <Card className="glass-card border-none overflow-hidden group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Teams</CardTitle>
              <div className="p-2 bg-primary/10 rounded-lg group-hover:bg-primary/20 transition-colors">
                <Users className="h-4 w-4 text-primary" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{stats?.totalTeams}</div>
              <p className="text-xs text-muted-foreground mt-1">Active groups you're in</p>
            </CardContent>
          </Card>
        </motion.div>
        
        <motion.div variants={item}>
          <Card className="glass-card border-none overflow-hidden group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">You Owe</CardTitle>
              <div className="p-2 bg-red-500/10 rounded-lg group-hover:bg-red-500/20 transition-colors">
                <ArrowDownRight className="h-4 w-4 text-red-500" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-red-500">${stats?.youOwe.toFixed(2)}</div>
              <p className="text-xs text-muted-foreground mt-1">Pending settlements</p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div variants={item}>
          <Card className="glass-card border-none overflow-hidden group">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">You are Owed</CardTitle>
              <div className="p-2 bg-green-500/10 rounded-lg group-hover:bg-green-500/20 transition-colors">
                <ArrowUpRight className="h-4 w-4 text-green-500" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-green-500">${stats?.youAreOwed.toFixed(2)}</div>
              <p className="text-xs text-muted-foreground mt-1">To be received</p>
            </CardContent>
          </Card>
        </motion.div>

        <motion.div variants={item}>
          <Card className="glass-card border-none overflow-hidden group bg-primary text-primary-foreground">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium opacity-80">Net Balance</CardTitle>
              <div className="p-2 bg-white/20 rounded-lg">
                <Wallet className="h-4 w-4 text-white" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">${netBalance.toFixed(2)}</div>
              <p className="text-xs opacity-70 mt-1">Overall financial status</p>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
        <motion.div variants={item} className="col-span-4">
          <Card className="glass-card border-none h-full">
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col items-center justify-center h-[300px] text-center">
              <div className="w-16 h-16 bg-secondary rounded-full flex items-center justify-center mb-4">
                <Receipt className="w-8 h-8 text-muted-foreground" />
              </div>
              <h3 className="font-semibold text-lg">No activity yet</h3>
              <p className="text-sm text-muted-foreground max-w-[250px]">
                When you join teams and add expenses, they'll show up here.
              </p>
            </CardContent>
          </Card>
        </motion.div>
        
        <motion.div variants={item} className="col-span-3">
          <Card className="glass-card border-none h-full">
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button className="w-full justify-between group h-12 rounded-xl" variant="outline">
                <div className="flex items-center">
                  <Users className="w-4 h-4 mr-3 text-primary" />
                  <span>Create New Team</span>
                </div>
                <Plus className="w-4 h-4 opacity-0 group-hover:opacity-100 transition-opacity" />
              </Button>
              <Button className="w-full justify-between group h-12 rounded-xl" variant="outline">
                <div className="flex items-center">
                  <Receipt className="w-4 h-4 mr-3 text-primary" />
                  <span>Add New Expense</span>
                </div>
                <Plus className="w-4 h-4 opacity-0 group-hover:opacity-100 transition-opacity" />
              </Button>
              <Button className="w-full justify-between group h-12 rounded-xl" variant="outline">
                <div className="flex items-center">
                  <TrendingUp className="w-4 h-4 mr-3 text-primary" />
                  <span>View Reports</span>
                </div>
                <ArrowUpRight className="w-4 h-4 opacity-0 group-hover:opacity-100 transition-opacity" />
              </Button>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </motion.div>
  );
}
