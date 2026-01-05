'use client';

import React, { useState, use } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Receipt, Users, BarChart3, Plus, Mail } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Skeleton } from '@/components/ui/skeleton';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

// Components
import { TeamHeader } from './components/TeamHeader';
import { ExpenseList } from './components/ExpenseList';
import { MemberList } from './components/MemberList';
import { BalanceSummary } from './components/BalanceSummary';
import { ApprovalList } from './components/ApprovalList';
import { AddExpenseDialog } from './components/AddExpenseDialog';

export default function TeamDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const queryClient = useQueryClient();
  const [isAddExpenseOpen, setIsAddExpenseOpen] = useState(false);
  const [isAddMemberOpen, setIsAddMemberOpen] = useState(false);
  const [newMemberEmail, setNewMemberEmail] = useState('');

  // Queries
  const { data: team, isLoading: isTeamLoading } = useQuery({
    queryKey: ['team', id],
    queryFn: async () => (await api.get(`/teams/${id}`)).data.data
  });

  const { data: expenses = [], isLoading: isExpensesLoading } = useQuery({
    queryKey: ['team-expenses', id],
    queryFn: async () => (await api.get(`/teams/${id}/expenses`)).data.data || []
  });

  const { data: members = [], isLoading: isMembersLoading } = useQuery({
    queryKey: ['team-members', id],
    queryFn: async () => (await api.get(`/teams/${id}/members`)).data.data || []
  });

  const { data: balancesData, isLoading: isBalancesLoading } = useQuery({
    queryKey: ['team-balances', id],
    queryFn: async () => (await api.get(`/teams/${id}/balances`)).data.data
  });

  const { data: approvals = [], isLoading: isApprovalsLoading } = useQuery({
    queryKey: ['team-approvals', id],
    queryFn: async () => {
      const res = await api.get(`/teams/${id}/approvals`);
      const approvalsData = res.data.data || [];
      return approvalsData.map((a: any) => ({
        ...a,
        expense: expenses.find((e: any) => e.id === a.expense_id)
      }));
    },
    enabled: !!expenses.length
  });

  // Mutations
  const addExpenseMutation = useMutation({
    mutationFn: async ({ data, file }: { data: any, file: File | null }) => {
      const amount = parseFloat(data.amount);
      const splitWith = data.split_with.length > 0 ? data.split_with : members.map((m: any) => m.user_id);
      
      const res = await api.post(`/teams/${id}/expenses`, {
        ...data,
        amount,
        split_with: splitWith,
      });
      
      if (file) {
        const formData = new FormData();
        formData.append('receipt', file);
        await api.post(`/teams/${id}/expenses/${res.data.data.id}/receipt`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
      }
      return res.data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-expenses', id] });
      queryClient.invalidateQueries({ queryKey: ['team-balances', id] });
      toast.success('Expense added successfully');
      setIsAddExpenseOpen(false);
    },
    onError: () => toast.error('Failed to add expense')
  });

  const deleteExpenseMutation = useMutation({
    mutationFn: async (expenseId: string) => api.delete(`/teams/${id}/expenses/${expenseId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-expenses', id] });
      queryClient.invalidateQueries({ queryKey: ['team-balances', id] });
      toast.success('Expense deleted');
    }
  });

  const addMemberMutation = useMutation({
    mutationFn: async (email: string) => api.post(`/teams/${id}/members`, { email, role: 'member' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', id] });
      toast.success('Member added successfully');
      setIsAddMemberOpen(false);
      setNewMemberEmail('');
    },
    onError: () => toast.error('Failed to add member')
  });

  const updateApprovalMutation = useMutation({
    mutationFn: async ({ approvalId, status }: { approvalId: string, status: string }) => 
      api.put(`/teams/${id}/approvals/${approvalId}`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-approvals', id] });
      queryClient.invalidateQueries({ queryKey: ['team-expenses', id] });
      toast.success('Approval updated');
    }
  });

  const settleMutation = useMutation({
    mutationFn: async ({ from, to, amount }: { from: string, to: string, amount: number }) => 
      api.post(`/teams/${id}/settlements`, { from_user: from, to_user: to, amount }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-balances', id] });
      toast.success('Settlement recorded');
    }
  });

  if (isTeamLoading) return (
    <div className="space-y-8">
      <Skeleton className="h-20 w-full rounded-2xl" />
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-6">
          <Skeleton className="h-12 w-64 rounded-xl" />
          <Skeleton className="h-[400px] w-full rounded-2xl" />
        </div>
        <div className="space-y-6">
          <Skeleton className="h-[300px] w-full rounded-2xl" />
          <Skeleton className="h-[300px] w-full rounded-2xl" />
        </div>
      </div>
    </div>
  );

  if (!team) return <div className="text-center py-20">Team not found</div>;

  return (
    <div className="space-y-8">
      <TeamHeader team={team} onAddExpense={() => setIsAddExpenseOpen(true)} />

      <ApprovalList 
        approvals={approvals} 
        onUpdate={(approvalId, status) => updateApprovalMutation.mutate({ approvalId, status })} 
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <Tabs defaultValue="expenses" className="w-full">
            <div className="flex items-center justify-between mb-6">
              <TabsList className="bg-secondary/50 p-1 rounded-xl">
                <TabsTrigger value="expenses" className="rounded-lg px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">
                  <Receipt className="w-4 h-4 mr-2" />
                  Expenses
                </TabsTrigger>
                <TabsTrigger value="members" className="rounded-lg px-6 data-[state=active]:bg-background data-[state=active]:shadow-sm">
                  <Users className="w-4 h-4 mr-2" />
                  Members
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="expenses" className="mt-0">
              <ExpenseList 
                expenses={expenses} 
                onDelete={(id) => deleteExpenseMutation.mutate(id)} 
              />
            </TabsContent>

            <TabsContent value="members" className="mt-0">
              <MemberList 
                members={members} 
                onAddMember={() => setIsAddMemberOpen(true)} 
              />
            </TabsContent>
          </Tabs>
        </div>

        <div className="space-y-8">
          <BalanceSummary 
            balances={balancesData?.balances || []} 
            memberBalances={balancesData?.members || []}
            onSettle={(from, to, amount) => settleMutation.mutate({ from, to, amount })}
          />
        </div>
      </div>

      <AddExpenseDialog 
        isOpen={isAddExpenseOpen} 
        onOpenChange={setIsAddExpenseOpen} 
        members={members}
        onSubmit={(data, file) => addExpenseMutation.mutate({ data, file })}
      />

      <Dialog open={isAddMemberOpen} onOpenChange={setIsAddMemberOpen}>
        <DialogContent className="sm:max-w-[425px] rounded-3xl">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold">Add Team Member</DialogTitle>
          </DialogHeader>
          <form onSubmit={(e) => { e.preventDefault(); addMemberMutation.mutate(newMemberEmail); }} className="space-y-6 py-4">
            <div className="grid gap-2">
              <Label htmlFor="email" className="text-sm font-semibold">Email Address</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input 
                  id="email" 
                  type="email" 
                  placeholder="colleague@example.com" 
                  className="pl-10 rounded-xl h-11"
                  value={newMemberEmail} 
                  onChange={e => setNewMemberEmail(e.target.value)} 
                  required 
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="submit" className="w-full rounded-xl h-11 shadow-lg shadow-primary/20">
                Send Invitation
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

