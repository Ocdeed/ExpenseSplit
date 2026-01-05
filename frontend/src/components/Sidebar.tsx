'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { LayoutDashboard, Receipt, Users, BarChart3, Settings, ChevronRight } from 'lucide-react';
import { motion } from 'framer-motion';

const menuItems = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Expenses', href: '/expenses', icon: Receipt },
  { name: 'Teams', href: '/teams', icon: Users },
  { name: 'Balances', href: '/balances', icon: BarChart3 },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed top-0 left-0 z-20 flex flex-col flex-shrink-0 w-64 h-full pt-16 font-normal duration-75 lg:flex transition-width" aria-label="Sidebar">
      <div className="relative flex flex-col flex-1 min-h-0 pt-0 bg-card border-r border-border/50">
        <div className="flex flex-col flex-1 pt-5 pb-4 overflow-y-auto">
          <div className="flex-1 px-4 space-y-1">
            <ul className="space-y-2">
              {menuItems.map((item) => {
                const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname.startsWith(item.href));
                return (
                  <li key={item.name}>
                    <Link
                      href={item.href}
                      className={cn(
                        "group relative flex items-center px-3 py-2.5 text-sm font-medium rounded-xl transition-all duration-200",
                        isActive 
                          ? "text-primary bg-secondary/80 shadow-sm" 
                          : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
                      )}
                    >
                      {isActive && (
                        <motion.div
                          layoutId="active-pill"
                          className="absolute left-0 w-1 h-6 bg-primary rounded-r-full"
                          transition={{ type: "spring", stiffness: 300, damping: 30 }}
                        />
                      )}
                      <item.icon className={cn(
                        "w-5 h-5 mr-3 transition-colors duration-200",
                        isActive ? "text-primary" : "group-hover:text-foreground"
                      )} />
                      <span className="flex-1">{item.name}</span>
                      {isActive && (
                        <ChevronRight className="w-4 h-4 text-primary/50" />
                      )}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        </div>
        
        <div className="p-4 border-t border-border/50">
          <Link
            href="/settings"
            className={cn(
              "flex items-center px-3 py-2.5 text-sm font-medium rounded-xl text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200",
              pathname === '/settings' && "text-primary bg-secondary/80"
            )}
          >
            <Settings className="w-5 h-5 mr-3" />
            <span>Settings</span>
          </Link>
        </div>
      </div>
    </aside>
  );
}
