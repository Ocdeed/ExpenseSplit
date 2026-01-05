'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Receipt, Plus, Info } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';

interface TeamMember {
  user_id: string;
  name: string;
  email: string;
}

interface AddExpenseDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  members: TeamMember[];
  onSubmit: (data: any, file: File | null) => void;
}

export function AddExpenseDialog({ isOpen, onOpenChange, members, onSubmit }: AddExpenseDialogProps) {
  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    category: 'General',
    split_type: 'equal',
    split_with: [] as string[],
    custom_splits: {} as Record<string, string>,
  });
  const [receiptFile, setReceiptFile] = useState<File | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(newExpense, receiptFile);
    setNewExpense({ description: '', amount: '', category: 'General', split_type: 'equal', split_with: [], custom_splits: {} });
    setReceiptFile(null);
  };

  const toggleMember = (userId: string) => {
    setNewExpense(prev => ({
      ...prev,
      split_with: prev.split_with.includes(userId)
        ? prev.split_with.filter(id => id !== userId)
        : [...prev.split_with, userId]
    }));
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px] rounded-3xl">
        <DialogHeader>
          <DialogTitle className="text-2xl font-bold">Add New Expense</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="desc" className="text-sm font-semibold">Description</Label>
              <Input 
                id="desc" 
                placeholder="What was it for?" 
                className="rounded-xl h-11"
                value={newExpense.description} 
                onChange={e => setNewExpense({...newExpense, description: e.target.value})} 
                required 
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="amount" className="text-sm font-semibold">Amount</Label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                  <Input 
                    id="amount" 
                    type="number" 
                    step="0.01" 
                    className="pl-7 rounded-xl h-11"
                    value={newExpense.amount} 
                    onChange={e => setNewExpense({...newExpense, amount: e.target.value})} 
                    required 
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="category" className="text-sm font-semibold">Category</Label>
                <Select value={newExpense.category} onValueChange={v => setNewExpense({...newExpense, category: v})}>
                  <SelectTrigger className="rounded-xl h-11">
                    <SelectValue placeholder="Select category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="General">General</SelectItem>
                    <SelectItem value="Food">Food</SelectItem>
                    <SelectItem value="Travel">Travel</SelectItem>
                    <SelectItem value="Entertainment">Entertainment</SelectItem>
                    <SelectItem value="Utilities">Utilities</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid gap-2">
              <Label className="text-sm font-semibold">Split With</Label>
              <ScrollArea className="h-[120px] rounded-xl border border-border/50 p-2">
                <div className="space-y-2">
                  {members.map(member => (
                    <div key={member.user_id} className="flex items-center space-x-2 p-1">
                      <Checkbox 
                        id={`member-${member.user_id}`} 
                        checked={newExpense.split_with.includes(member.user_id)}
                        onCheckedChange={() => toggleMember(member.user_id)}
                      />
                      <label htmlFor={`member-${member.user_id}`} className="text-sm font-medium leading-none cursor-pointer">
                        {member.name}
                      </label>
                    </div>
                  ))}
                </div>
              </ScrollArea>
              <p className="text-[10px] text-muted-foreground flex items-center gap-1">
                <Info className="w-3 h-3" />
                Leave empty to split with everyone
              </p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="receipt" className="text-sm font-semibold">Receipt (Optional)</Label>
              <div className="flex items-center gap-2">
                <Input 
                  id="receipt" 
                  type="file" 
                  accept="image/*" 
                  className="hidden" 
                  onChange={e => setReceiptFile(e.target.files?.[0] || null)} 
                />
                <Button 
                  type="button" 
                  variant="outline" 
                  className="w-full rounded-xl h-11 border-dashed border-2 hover:border-primary hover:bg-primary/5 transition-all"
                  onClick={() => document.getElementById('receipt')?.click()}
                >
                  <Receipt className="w-4 h-4 mr-2" />
                  {receiptFile ? receiptFile.name : 'Upload Receipt'}
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="submit" className="w-full rounded-xl h-11 shadow-lg shadow-primary/20">
              Create Expense
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
