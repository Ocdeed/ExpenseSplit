'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Plus, UserPlus, Mail, Shield } from 'lucide-react';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/badge';

interface TeamMember {
  user_id: string;
  name: string;
  email: string;
  role: string;
}

interface MemberListProps {
  members: TeamMember[];
  onAddMember: () => void;
}

export function MemberList({ members, onAddMember }: MemberListProps) {
  return (
    <Card className="glass-card border-none">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-lg font-bold flex items-center gap-2">
          <UserPlus className="w-5 h-5 text-primary" />
          Team Members
        </CardTitle>
        <Button variant="ghost" size="sm" className="rounded-full h-8 w-8 p-0" onClick={onAddMember}>
          <Plus className="w-4 h-4" />
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {members.map((member, idx) => (
          <motion.div
            key={member.user_id}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: idx * 0.05 }}
            className="flex items-center justify-between p-3 rounded-xl bg-secondary/30 hover:bg-secondary/50 transition-colors group"
          >
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center font-bold text-primary">
                {member.name[0].toUpperCase()}
              </div>
              <div>
                <div className="font-semibold text-sm flex items-center gap-2">
                  {member.name}
                  {member.role === 'admin' && (
                    <Badge variant="outline" className="text-[10px] h-4 px-1 bg-primary/5 text-primary border-primary/20">
                      Admin
                    </Badge>
                  )}
                </div>
                <div className="text-xs text-muted-foreground flex items-center gap-1">
                  <Mail className="w-3 h-3" />
                  {member.email}
                </div>
              </div>
            </div>
          </motion.div>
        ))}
      </CardContent>
    </Card>
  );
}
