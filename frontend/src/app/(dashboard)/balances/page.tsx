'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3, TrendingUp, TrendingDown } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';

interface TeamBalance {
  team_id: string;
  team_name: string;
  balance: number;
}

export default function BalancesPage() {
  const [teamBalances, setTeamBalances] = useState<TeamBalance[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [totalOwe, setTotalOwe] = useState(0);
  const [totalOwed, setTotalOwed] = useState(0);

  useEffect(() => {
    const fetchBalances = async () => {
      try {
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

        setTeamBalances(balances);
        setTotalOwe(owe);
        setTotalOwed(owed);
      } catch (error) {
        toast.error('Failed to fetch balances');
      } finally {
        setIsLoading(false);
      }
    };

    fetchBalances();
  }, []);

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Your Balances</h1>

      <div className="grid gap-4 md:grid-cols-2">
        <Card className="bg-red-50 border-red-100">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-red-800">Total You Owe</CardTitle>
            <TrendingDown className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-700">${totalOwe.toFixed(2)}</div>
          </CardContent>
        </Card>
        <Card className="bg-green-50 border-green-100">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-green-800">Total You Are Owed</CardTitle>
            <TrendingUp className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-700">${totalOwed.toFixed(2)}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <BarChart3 className="w-5 h-5 mr-2" />
            Balances by Team
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <div className="space-y-4">
              {teamBalances.length === 0 ? (
                <p className="text-center py-8 text-gray-500">No balances found</p>
              ) : (
                teamBalances.map((tb) => (
                  <div key={tb.team_id} className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors">
                    <div>
                      <p className="font-medium text-gray-900">{tb.team_name}</p>
                      <p className="text-sm text-gray-500">Team ID: {tb.team_id}</p>
                    </div>
                    <div className={`text-lg font-bold ${tb.balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {tb.balance >= 0 ? `+ $${tb.balance.toFixed(2)}` : `- $${Math.abs(tb.balance).toFixed(2)}`}
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
