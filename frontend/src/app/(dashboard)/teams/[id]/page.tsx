'use client';

import React, { useEffect, useState, use } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Plus, Receipt, Users, BarChart3, Download, Trash2, CheckCircle, XCircle, Clock, FileText } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';

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

interface Approval {
  id: string;
  expense_id: string;
  status: 'pending' | 'approved' | 'rejected';
  comment?: string;
  created_at: string;
  expense?: Expense;
}

interface TeamMember {
  user_id: string;
  name: string;
  email: string;
  role: string;
}

interface Team {
  id: string;
  name: string;
  description: string;
}

export default function TeamDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [team, setTeam] = useState<Team | null>(null);
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [balances, setBalances] = useState<any[]>([]);
  const [memberBalances, setMemberBalances] = useState<any[]>([]);
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  
  const [isAddExpenseOpen, setIsAddExpenseOpen] = useState(false);
  const [isAddMemberOpen, setIsAddMemberOpen] = useState(false);
  const [newMemberEmail, setNewMemberEmail] = useState('');
  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    category: 'General',
    split_type: 'equal',
    split_with: [] as string[],
  });
  const [receiptFile, setReceiptFile] = useState<File | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [teamRes, expensesRes, membersRes, balancesRes, approvalsRes] = await Promise.all([
          api.get(`/teams/${id}`),
          api.get(`/teams/${id}/expenses`),
          api.get(`/teams/${id}/members`),
          api.get(`/teams/${id}/balances`),
          api.get(`/teams/${id}/approvals`),
        ]);
        
        setTeam(teamRes.data.data);
        const expensesData = expensesRes.data.data || [];
        setExpenses(expensesData);
        setMembers(membersRes.data.data || []);
        setBalances(balancesRes.data.data.balances || []);
        setMemberBalances(balancesRes.data.data.members || []);
        
        // Enrich approvals with expense info
        const enrichedApprovals = (approvalsRes.data.data || []).map((a: any) => {
          const exp = expensesData.find((e: any) => e.id === a.expense_id);
          return {
            ...a,
            expense: exp,
          };
        });
        setApprovals(enrichedApprovals);
      } catch (error) {
        console.error(error);
        toast.error('Failed to fetch team data');
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [id]);

  const refreshData = async () => {
    try {
      const [expensesRes, balancesRes, approvalsRes] = await Promise.all([
        api.get(`/teams/${id}/expenses`),
        api.get(`/teams/${id}/balances`),
        api.get(`/teams/${id}/approvals`),
      ]);
      
      const expensesData = expensesRes.data.data || [];
      setExpenses(expensesData);
      setBalances(balancesRes.data.data.balances || []);
      setMemberBalances(balancesRes.data.data.members || []);
      
      const enrichedApprovals = (approvalsRes.data.data || []).map((a: any) => {
        const exp = expensesData.find((e: any) => e.id === a.expense_id);
        return { ...a, expense: exp };
      });
      setApprovals(enrichedApprovals);
    } catch (error) {
      console.error('Failed to refresh data', error);
    }
  };

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post(`/teams/${id}/members`, {
        email: newMemberEmail,
        role: 'member',
      });
      toast.success('Member added successfully');
      setIsAddMemberOpen(false);
      setNewMemberEmail('');
      
      // Refresh members
      const membersRes = await api.get(`/teams/${id}/members`);
      setMembers(membersRes.data.data || []);
    } catch (error) {
      toast.error('Failed to add member. Make sure the user exists.');
    }
  };

  const handleAddExpense = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const amount = parseFloat(newExpense.amount);
      if (isNaN(amount) || amount <= 0) {
        toast.error('Please enter a valid amount');
        return;
      }
      
      const splitWith = newExpense.split_with.length > 0 ? newExpense.split_with : members.map(m => m.user_id);
      if (splitWith.length === 0) {
        toast.error('Please select at least one member to split with');
        return;
      }

      const response = await api.post(`/teams/${id}/expenses`, {
        description: newExpense.description,
        amount: amount,
        category: newExpense.category,
        split_type: newExpense.split_type,
        split_with: splitWith,
      });
      
      const expenseId = response.data.data.id;

      // Upload receipt if exists
      if (receiptFile) {
        const formData = new FormData();
        formData.append('receipt', receiptFile);
        await api.post(`/teams/${id}/expenses/${expenseId}/receipt`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
      }

      toast.success('Expense added successfully');
      setIsAddExpenseOpen(false);
      setNewExpense({ description: '', amount: '', category: 'General', split_type: 'equal', split_with: [] });
      setReceiptFile(null);
      
      refreshData();
    } catch (error) {
      toast.error('Failed to add expense');
    }
  };

  const handleUpdateApproval = async (approvalId: string, status: 'approved' | 'rejected') => {
    try {
      await api.put(`/teams/${id}/approvals/${approvalId}`, { status });
      toast.success(`Expense ${status}`);
      refreshData();
    } catch (error) {
      toast.error('Failed to update approval');
    }
  };

  const handleSettle = async (fromUser: string, toUser: string, amount: number) => {
    try {
      await api.post(`/teams/${id}/settlements`, {
        from_user: fromUser,
        to_user: toUser,
        amount: amount,
      });
      toast.success('Settlement recorded');
      refreshData();
    } catch (error) {
      toast.error('Failed to record settlement');
    }
  };

  const handleDeleteExpense = async (expenseId: string) => {
    if (!confirm('Are you sure you want to delete this expense?')) return;
    try {
      await api.delete(`/teams/${id}/expenses/${expenseId}`);
      toast.success('Expense deleted');
      refreshData();
    } catch (error) {
      toast.error('Failed to delete expense');
    }
  };

  const handleExport = async (type: string) => {
    try {
      const response = await api.get(`/teams/${id}/export/${type}`, {
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `${type}_${id}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      toast.error('Failed to export data');
    }
  };

  if (isLoading) return <div className="flex justify-center py-12"><div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div></div>;
  if (!team) return <div>Team not found</div>;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{team.name}</h1>
          <p className="text-gray-500">{team.description}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => handleExport('summary')}>
            <Download className="w-4 h-4 mr-2" />
            Summary Report
          </Button>
          <Button variant="outline" onClick={() => handleExport('expenses')}>
            <Download className="w-4 h-4 mr-2" />
            Export CSV
          </Button>
          <Dialog open={isAddExpenseOpen} onOpenChange={setIsAddExpenseOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                Add Expense
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add New Expense</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleAddExpense}>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="desc">Description</Label>
                    <Input id="desc" value={newExpense.description} onChange={e => setNewExpense({...newExpense, description: e.target.value})} required />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="amount">Amount</Label>
                    <Input id="amount" type="number" step="0.01" value={newExpense.amount} onChange={e => setNewExpense({...newExpense, amount: e.target.value})} required />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="split_type">Split Type</Label>
                    <Select value={newExpense.split_type} onValueChange={v => setNewExpense({...newExpense, split_type: v})}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select split type" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="equal">Equal Split</SelectItem>
                        <SelectItem value="custom">Custom Amount</SelectItem>
                        <SelectItem value="percent">Percentage</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="category">Category</Label>
                    <Select value={newExpense.category} onValueChange={v => setNewExpense({...newExpense, category: v})}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select category" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="General">General</SelectItem>
                        <SelectItem value="Food & Dining">Food & Dining</SelectItem>
                        <SelectItem value="Transportation">Transportation</SelectItem>
                        <SelectItem value="Office Supplies">Office Supplies</SelectItem>
                        <SelectItem value="Software & Tools">Software & Tools</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label>Split With</Label>
                    <div className="grid grid-cols-2 gap-2 border p-2 rounded-md max-h-32 overflow-y-auto">
                      {members.map(member => (
                        <div key={member.user_id} className="flex items-center space-x-2">
                          <Checkbox 
                            id={`member-${member.user_id}`} 
                            checked={newExpense.split_with.includes(member.user_id)}
                            onCheckedChange={(checked) => {
                              if (checked) {
                                setNewExpense({...newExpense, split_with: [...newExpense.split_with, member.user_id]});
                              } else {
                                setNewExpense({...newExpense, split_with: newExpense.split_with.filter(id => id !== member.user_id)});
                              }
                            }}
                          />
                          <label htmlFor={`member-${member.user_id}`} className="text-sm truncate">{member.name}</label>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="receipt">Receipt (Optional)</Label>
                    <Input id="receipt" type="file" onChange={e => setReceiptFile(e.target.files?.[0] || null)} />
                  </div>
                </div>
                <DialogFooter>
                  <Button type="submit">Add Expense</Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <Tabs defaultValue="expenses">
        <TabsList>
          <TabsTrigger value="expenses">
            <Receipt className="w-4 h-4 mr-2" />
            Expenses
          </TabsTrigger>
          <TabsTrigger value="balances">
            <BarChart3 className="w-4 h-4 mr-2" />
            Balances
          </TabsTrigger>
          <TabsTrigger value="approvals">
            <CheckCircle className="w-4 h-4 mr-2" />
            Approvals
          </TabsTrigger>
          <TabsTrigger value="members">
            <Users className="w-4 h-4 mr-2" />
            Members
          </TabsTrigger>
        </TabsList>

        <TabsContent value="expenses" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Expense History</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Payer</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {expenses.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7} className="text-center py-8 text-gray-500">No expenses yet</TableCell>
                    </TableRow>
                  ) : (
                    expenses.map((expense) => (
                      <TableRow key={expense.id}>
                        <TableCell>{new Date(expense.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="font-medium">
                          <div className="flex items-center gap-2">
                            {expense.description}
                            {expense.receipt_url && (
                              <a href={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}${expense.receipt_url}`} target="_blank" rel="noreferrer">
                                <FileText className="w-4 h-4 text-blue-500 cursor-pointer" />
                              </a>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>{expense.paid_by.name}</TableCell>
                        <TableCell>{expense.category}</TableCell>
                        <TableCell>
                          <Badge variant={
                            expense.approval_status === 'approved' ? 'default' : 
                            expense.approval_status === 'rejected' ? 'destructive' : 
                            'outline'
                          }>
                            {expense.approval_status}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right font-bold">${expense.amount.toFixed(2)}</TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="sm" onClick={() => handleDeleteExpense(expense.id)}>
                            <Trash2 className="w-4 h-4 text-red-500" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="balances" className="mt-6">
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Settlements Needed</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {balances.length === 0 ? (
                    <p className="text-center py-4 text-gray-500">All settled up!</p>
                  ) : (
                    balances.map((balance, i) => (
                      <div key={i} className="flex items-center justify-between p-4 border rounded-lg">
                        <div className="text-sm">
                          <span className="font-bold">{balance.from_user.name}</span> owes <span className="font-bold">{balance.to_user.name}</span>
                        </div>
                        <div className="flex items-center gap-3">
                          <div className="text-lg font-bold text-red-600">
                            ${balance.amount.toFixed(2)}
                          </div>
                          <Button size="sm" onClick={() => handleSettle(balance.from_user.id, balance.to_user.id, balance.amount)}>
                            Settle
                          </Button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Member Summaries</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {memberBalances.map((mb) => (
                    <div key={mb.user.id} className="flex items-center justify-between p-4 border rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold">
                          {mb.user.name.charAt(0)}
                        </div>
                        <div>
                          <p className="font-medium">{mb.user.name}</p>
                          <p className="text-xs text-gray-500">Owes: ${mb.total_owed.toFixed(2)} | Owed: ${mb.total_owing.toFixed(2)}</p>
                        </div>
                      </div>
                      <div className={`text-lg font-bold ${mb.net_balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {mb.net_balance >= 0 ? `+ $${mb.net_balance.toFixed(2)}` : `- $${Math.abs(mb.net_balance).toFixed(2)}`}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="approvals" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Expense Approvals</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Expense</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {approvals.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-gray-500">No approvals found</TableCell>
                    </TableRow>
                  ) : (
                    approvals.map((approval) => (
                      <TableRow key={approval.id}>
                        <TableCell className="font-medium">{approval.expense?.description || 'Unknown Expense'}</TableCell>
                        <TableCell>${approval.expense?.amount?.toFixed(2) || '0.00'}</TableCell>
                        <TableCell>
                          <Badge variant={approval.status === 'approved' ? 'default' : approval.status === 'rejected' ? 'destructive' : 'outline'}>
                            {approval.status}
                          </Badge>
                        </TableCell>
                        <TableCell>{new Date(approval.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right">
                          {approval.status === 'pending' && (
                            <div className="flex justify-end gap-2">
                              <Button size="sm" variant="outline" className="text-green-600" onClick={() => handleUpdateApproval(approval.id, 'approved')}>
                                <CheckCircle className="w-4 h-4 mr-1" /> Approve
                              </Button>
                              <Button size="sm" variant="outline" className="text-red-600" onClick={() => handleUpdateApproval(approval.id, 'rejected')}>
                                <XCircle className="w-4 h-4 mr-1" /> Reject
                              </Button>
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="members" className="mt-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Team Members</CardTitle>
              <Dialog open={isAddMemberOpen} onOpenChange={setIsAddMemberOpen}>
                <DialogTrigger asChild>
                  <Button size="sm">
                    <Plus className="w-4 h-4 mr-2" />
                    Add Member
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Add Team Member</DialogTitle>
                  </DialogHeader>
                  <form onSubmit={handleAddMember}>
                    <div className="grid gap-4 py-4">
                      <div className="grid gap-2">
                        <Label htmlFor="email">User Email</Label>
                        <Input 
                          id="email" 
                          type="email" 
                          placeholder="user@example.com" 
                          value={newMemberEmail} 
                          onChange={e => setNewMemberEmail(e.target.value)} 
                          required 
                        />
                      </div>
                    </div>
                    <DialogFooter>
                      <Button type="submit">Add Member</Button>
                    </DialogFooter>
                  </form>
                </DialogContent>
              </Dialog>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {members.map((member) => (
                  <div key={member.user_id} className="flex items-center justify-between p-4 border rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center text-gray-600 font-bold">
                        {member.name.charAt(0)}
                      </div>
                      <div>
                        <p className="font-medium">{member.name}</p>
                        <p className="text-sm text-gray-500">{member.email}</p>
                      </div>
                    </div>
                    <Badge variant="secondary">{member.role}</Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

