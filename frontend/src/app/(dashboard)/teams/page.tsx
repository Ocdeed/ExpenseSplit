'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Plus, Users, ArrowRight, Search, LayoutGrid, List } from 'lucide-react';
import api from '@/lib/api';
import { toast } from 'sonner';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'framer-motion';
import { Skeleton } from '@/components/ui/skeleton';

interface Team {
  id: string;
  name: string;
  description: string;
  owner_id: string;
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.1 }
  }
};

const item = {
  hidden: { opacity: 0, scale: 0.95 },
  show: { opacity: 1, scale: 1 }
};

export default function TeamsPage() {
  const queryClient = useQueryClient();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [newTeamName, setNewTeamName] = useState('');
  const [newTeamDescription, setNewTeamDescription] = useState('');
  const [searchTerm, setSearchTerm] = useState('');

  const { data: teams = [], isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: async () => (await api.get('/teams')).data.data || []
  });

  const createTeamMutation = useMutation({
    mutationFn: async (data: { name: string, description: string }) => api.post('/teams', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      toast.success('Team created successfully');
      setIsCreateDialogOpen(false);
      setNewTeamName('');
      setNewTeamDescription('');
    },
    onError: () => toast.error('Failed to create team')
  });

  const filteredTeams = teams.filter((t: Team) => 
    t.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    t.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="space-y-8"
    >
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tight text-gradient">Your Teams</h1>
          <p className="text-muted-foreground mt-1">Manage your groups and shared expenses.</p>
        </div>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="rounded-full px-6 shadow-lg shadow-primary/20">
              <Plus className="w-4 h-4 mr-2" />
              Create Team
            </Button>
          </DialogTrigger>
          <DialogContent className="rounded-3xl">
            <DialogHeader>
              <DialogTitle className="text-2xl font-bold">Create New Team</DialogTitle>
              <DialogDescription>
                Create a team to start splitting expenses with others.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={(e) => { e.preventDefault(); createTeamMutation.mutate({ name: newTeamName, description: newTeamDescription }); }}>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="name" className="font-semibold">Team Name</Label>
                  <Input
                    id="name"
                    value={newTeamName}
                    onChange={(e) => setNewTeamName(e.target.value)}
                    placeholder="e.g. Trip to Paris"
                    className="rounded-xl h-11"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="description" className="font-semibold">Description</Label>
                  <Input
                    id="description"
                    value={newTeamDescription}
                    onChange={(e) => setNewTeamDescription(e.target.value)}
                    placeholder="Optional description"
                    className="rounded-xl h-11"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button type="submit" className="w-full rounded-xl h-11 shadow-lg shadow-primary/20">
                  Create Team
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="relative w-full md:w-96 group">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
        <Input
          placeholder="Search teams..."
          className="pl-10 bg-card/50 border-border/50 rounded-xl h-11 focus-visible:ring-primary/20"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      {isLoading ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-48 w-full rounded-2xl" />
          ))}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <AnimatePresence mode="popLayout">
            {filteredTeams.length === 0 ? (
              <div className="col-span-full py-20 text-center">
                <div className="flex flex-col items-center justify-center text-muted-foreground">
                  <Users className="w-12 h-12 mb-4 opacity-20" />
                  <p className="text-lg font-medium">No teams found</p>
                  <p className="text-sm">Create a new team to get started!</p>
                </div>
              </div>
            ) : (
              filteredTeams.map((team: Team) => (
                <motion.div key={team.id} variants={item} layout>
                  <Link href={`/teams/${team.id}`}>
                    <Card className="glass-card border-none h-full group cursor-pointer overflow-hidden">
                      <CardHeader>
                        <div className="w-12 h-12 bg-primary/10 rounded-2xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-300">
                          <Users className="w-6 h-6 text-primary" />
                        </div>
                        <CardTitle className="text-xl group-hover:text-primary transition-colors">{team.name}</CardTitle>
                        <CardDescription className="line-clamp-2">{team.description || 'No description provided.'}</CardDescription>
                      </CardHeader>
                      <CardFooter className="border-t border-border/50 bg-secondary/30 py-3 flex justify-between items-center">
                        <span className="text-xs font-medium text-muted-foreground">View Details</span>
                        <ArrowRight className="w-4 h-4 text-primary opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all" />
                      </CardFooter>
                    </Card>
                  </Link>
                </motion.div>
              ))
            )}
          </AnimatePresence>
        </div>
      )}
    </motion.div>
  );
}
