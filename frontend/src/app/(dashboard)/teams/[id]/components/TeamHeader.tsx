'use client';

import { Button } from '@/components/ui/button';
import { Download, Plus } from 'lucide-react';
import { motion } from 'framer-motion';
import api from '@/lib/api';
import { toast } from 'sonner';

interface TeamHeaderProps {
  team: {
    id: string;
    name: string;
    description: string;
  };
  onAddExpense: () => void;
}

export function TeamHeader({ team, onAddExpense }: TeamHeaderProps) {
  const handleExport = async (type: string) => {
    try {
      const response = await api.get(`/teams/${team.id}/export/${type}`, {
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `${type}_${team.id}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      toast.error('Failed to export data');
    }
  };

  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
      >
        <h1 className="text-4xl font-bold tracking-tight text-gradient">{team.name}</h1>
        <p className="text-muted-foreground mt-1">{team.description}</p>
      </motion.div>
      
      <motion.div 
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        className="flex flex-wrap gap-2"
      >
        <Button variant="outline" className="rounded-full" onClick={() => handleExport('summary')}>
          <Download className="w-4 h-4 mr-2" />
          Summary
        </Button>
        <Button variant="outline" className="rounded-full" onClick={() => handleExport('expenses')}>
          <Download className="w-4 h-4 mr-2" />
          CSV
        </Button>
        <Button className="rounded-full px-6 shadow-lg shadow-primary/20" onClick={onAddExpense}>
          <Plus className="w-4 h-4 mr-2" />
          Add Expense
        </Button>
      </motion.div>
    </div>
  );
}
