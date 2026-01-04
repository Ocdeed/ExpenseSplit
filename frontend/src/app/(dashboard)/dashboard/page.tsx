'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Receipt, Users, TrendingUp, TrendingDown } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';

interface DashboardStats {
  totalExpenses: number;
  totalTeams: number;
  youOwe: number;
  youAreOwed: number;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats>({
    totalExpenses: 0,
    totalTeams: 0,
    youOwe: 0,
    youAreOwed: 0,
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        // In a real app, you might have a dedicated stats endpoint
        // For now, we'll fetch teams and some balances to simulate
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

        setStats({
          totalExpenses: 0, // Would need another call or backend support
          totalTeams: teams.length,
          youOwe: totalOwe,
          youAreOwed: totalOwed,
        });
      } catch (error) {
        console.error('Failed to fetch dashboard stats', error);
        toast.error('Failed to load dashboard data');
      } finally {
        setIsLoading(false);
      }
    };

    fetchStats();
  }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
      
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Teams</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalTeams}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">You Owe</CardTitle>
            <TrendingDown className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">${stats.youOwe.toFixed(2)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">You are Owed</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">${stats.youAreOwed.toFixed(2)}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Balance</CardTitle>
            <Receipt className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${(stats.youAreOwed - stats.youOwe).toFixed(2)}</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">No recent activity found.</p>
          </CardContent>
        </Card>
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button className="w-full justify-start" variant="outline">Create New Team</Button>
            <Button className="w-full justify-start" variant="outline">Add Expense</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
