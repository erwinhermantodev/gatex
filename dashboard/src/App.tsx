import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { 
  LayoutDashboard, 
  Globe, 
  Server, 
  Braces, 
  Settings, 
  Plus, 
  Trash2, 
  Edit3, 
  Activity,
  CheckCircle2,
  Bell,
  Search,
  ChevronDown,
  Clock,
  Zap,
  Shield,
  RefreshCw,
  MoreVertical
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface Service {
  ID: number;
  Name: string;
  BaseURL: string;
  Protocol: string;
  GRPCAddr: string;
}

interface Route {
  ID: number;
  Path: string;
  Method: string;
  ServiceID: number;
  Service?: Service;
  EndpointFilter: string;
  Tag: string;
}

interface ProtoMapping {
  ID: number;
  ServiceID: number;
  Service?: Service;
  RPCMethod: string;
  ProtoPackage: string;
  RequestType: string;
  ResponseType: string;
}

const TABS = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'services', label: 'Services', icon: Server },
  { id: 'routes', label: 'Routes', icon: Globe },
  { id: 'proto', label: 'Proto Mappings', icon: Braces },
  { id: 'settings', label: 'Settings', icon: Settings },
];

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [services, setServices] = useState<Service[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [protoMappings, setProtoMappings] = useState<ProtoMapping[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchData();
  }, [activeTab]);

  const fetchData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'services' || activeTab === 'dashboard') {
        const sRes = await axios.get('/admin/services');
        setServices(sRes.data || []);
      }
      if (activeTab === 'routes' || activeTab === 'dashboard') {
        const rRes = await axios.get('/admin/routes');
        setRoutes(rRes.data || []);
      }
      if (activeTab === 'proto' || activeTab === 'dashboard') {
        const pRes = await axios.get('/admin/proto-mappings');
        setProtoMappings(pRes.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch data', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex h-screen bg-[#060608] text-[#e1e1e3] font-sans selection:bg-cyan-500/30 overflow-hidden">
      {/* Sidebar */}
      <aside className="w-[280px] bg-[#0d0d0f] border-r border-white/5 flex flex-col">
        <div className="p-6">
          <div className="flex items-center justify-between p-3 bg-white/5 rounded-2xl border border-white/10 group cursor-pointer hover:bg-white/10 transition-all">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-cyan-500 to-blue-600 rounded-xl flex items-center justify-center shadow-[0_0_20px_rgba(6,182,212,0.3)]">
                <Activity size={24} className="text-white" />
              </div>
              <div className="flex flex-col">
                <span className="font-bold text-sm tracking-tight">Antigravity</span>
                <span className="text-[10px] text-zinc-500 font-medium uppercase tracking-widest">Main Gateway</span>
              </div>
            </div>
            <ChevronDown size={16} className="text-zinc-500 group-hover:text-zinc-300 transition-colors" />
          </div>
        </div>

        <nav className="flex-1 px-4 py-2 space-y-1">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-300 group ${
                activeTab === tab.id 
                ? 'bg-cyan-500/10 text-cyan-400 font-semibold' 
                : 'text-zinc-500 hover:text-zinc-200 hover:bg-white/5'
              }`}
            >
              <tab.icon size={20} className={activeTab === tab.id ? 'text-cyan-400' : 'text-zinc-500 group-hover:text-zinc-300'} />
              <span className="text-[14px]">{tab.label}</span>
              {activeTab === tab.id && (
                <motion.div layoutId="activeTab" className="ml-auto w-1.5 h-1.5 rounded-full bg-cyan-400 shadow-[0_0_10px_rgba(6,182,212,0.5)]" />
              )}
            </button>
          ))}
        </nav>

        <div className="p-6 mt-auto border-t border-white/5 bg-gradient-to-t from-black/20">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-zinc-700 to-zinc-800 border border-white/10 flex items-center justify-center overflow-hidden">
               <span className="text-zinc-400 font-bold text-xs uppercase">AD</span>
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-bold">Admin User</span>
              <span className="text-[10px] text-zinc-500 font-medium">admin@antigravity.io</span>
            </div>
            <MoreVertical size={16} className="ml-auto text-zinc-600 hover:text-zinc-400 cursor-pointer" />
          </div>
        </div>
      </aside>

      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-[80px] border-b border-white/5 px-10 flex items-center justify-between bg-black/10 backdrop-blur-md">
          <div className="flex flex-col">
             <div className="flex items-center gap-2 text-xs text-zinc-500 font-medium tracking-tight mb-1">
                <span>Infrastructure</span>
                <span className="w-1 h-1 rounded-full bg-zinc-700" />
                <span className="text-zinc-400">Gateway Dashboard</span>
             </div>
             <h2 className="text-2xl font-bold tracking-tight capitalize">{activeTab === 'dashboard' ? 'System Overview' : activeTab}</h2>
          </div>

          <div className="flex items-center gap-6">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
              <input 
                type="text" 
                placeholder="Search routes..." 
                className="bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-sm w-[240px] focus:outline-none focus:border-cyan-500/50 transition-all"
              />
            </div>
            
            <button className="relative text-zinc-400 hover:text-zinc-200 transition-colors">
              <Bell size={20} />
              <span className="absolute top-0 right-0 w-2 h-2 bg-rose-500 rounded-full border-2 border-[#060608]" />
            </button>

            <div className="flex items-center gap-2 px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-full">
              <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <span className="text-[11px] font-bold text-emerald-500 uppercase tracking-wider">Gateway Operational</span>
            </div>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-10 space-y-8 bg-[#08080a]">
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.3, ease: 'easeOut' }}
            >
              {activeTab === 'dashboard' && <DashboardContent services={services} routes={routes} onRefresh={fetchData} />}
              {activeTab === 'services' && <ServicesModule services={services} onRefresh={fetchData} />}
              {activeTab === 'routes' && <RoutesModule routes={routes} services={services} onRefresh={fetchData} />}
              {activeTab === 'proto' && <ProtoMappingsModule mappings={protoMappings} onRefresh={fetchData} />}
            </motion.div>
          </AnimatePresence>
        </main>
      </div>
    </div>
  );
}

function DashboardContent({ services, routes, onRefresh }: { services: Service[], routes: Route[], onRefresh: () => void }) {
  return (
    <div className="grid grid-cols-12 gap-8">
      <div className="col-span-12 grid grid-cols-4 gap-6">
        <StatCard label="Active Services" value={services.length} trend="+2%" status="Stable" icon={Server} color="cyan" />
        <StatCard label="Configured Routes" value={routes.length} trend="+12" status="Active" icon={Globe} color="purple" />
        <StatCard label="Gateway Health" value="99.9%" trend="Uptime" status="Excellent" icon={Shield} color="emerald" />
        <StatCard label="Avg. Latency" value="24ms" trend="-4%" status="Stable" icon={Zap} color="orange" />
      </div>

      <div className="col-span-8 bg-zinc-900/40 rounded-3xl border border-white/5 p-8 relative overflow-hidden">
        <div className="flex justify-between items-center mb-10">
          <div>
            <h3 className="text-lg font-bold">Live Traffic Monitor</h3>
            <p className="text-xs text-zinc-500 font-medium">Throughput across all endpoints (Last 60 Minutes)</p>
          </div>
          <button onClick={onRefresh} className="p-2 bg-white/5 hover:bg-white/10 rounded-xl transition-all">
            <RefreshCw size={18} className="text-zinc-400" />
          </button>
        </div>
        <div className="h-[280px] flex items-end gap-3 px-4">
          {[40, 70, 45, 90, 65, 80, 50, 60, 100, 85, 75, 45, 90, 65, 80, 50, 60, 40].map((h, i) => (
            <div key={i} className="flex-1 bg-white/5 rounded-t-lg relative group transition-all" style={{ height: `${h}%` }}>
              <div className="absolute inset-0 bg-gradient-to-t from-cyan-600/20 to-cyan-400 scale-y-0 group-hover:scale-y-100 transition-transform origin-bottom rounded-t-lg" />
            </div>
          ))}
        </div>
      </div>

      <div className="col-span-4 bg-zinc-900/40 rounded-3xl border border-white/5 p-8 flex flex-col">
        <h3 className="text-lg font-bold mb-8">Recent Activity</h3>
        <div className="space-y-8 flex-1">
          {[
            { msg: 'Route /v1/auth updated', time: '2m ago', by: 'admin', color: 'bg-cyan-500' },
            { msg: 'Service payment-svc added', time: '1h ago', by: 'system', color: 'bg-indigo-500' },
            { msg: 'System reload successful', time: '3h ago', by: 'system', color: 'bg-emerald-500' },
            { msg: 'New API key generated', time: 'Yesterday', by: 'user_42', color: 'bg-purple-500' }
          ].map((item, i) => (
            <div key={i} className="flex gap-4 relative">
              {i < 3 && <div className="absolute left-[3px] top-6 w-0.5 h-10 bg-white/5" />}
              <div className={`w-2 h-2 rounded-full ${item.color} mt-1.5 shadow-[0_0_8px_rgba(255,255,255,0.2)]`} />
              <div className="flex flex-col gap-1">
                <span className="text-sm font-semibold text-zinc-200">{item.msg}</span>
                <span className="text-[10px] text-zinc-500 font-medium">By {item.by} • {item.time}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="col-span-12 bg-zinc-900/40 rounded-3xl border border-white/5 p-8">
        <h3 className="text-lg font-bold mb-8">Priority Routes</h3>
        <div className="overflow-hidden">
          <table className="w-full text-left font-sans">
            <thead>
              <tr className="text-[10px] text-zinc-500 uppercase tracking-widest border-b border-white/5">
                <th className="pb-6">Path</th>
                <th className="pb-6">Method</th>
                <th className="pb-6">Upstream Service</th>
                <th className="pb-6">Status</th>
                <th className="pb-6 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {routes.slice(0, 5).map((r) => (
                <tr key={r.ID} className="group hover:bg-white/[0.02] transition-colors">
                  <td className="py-6 font-mono text-xs text-cyan-400/80">{r.Path}</td>
                  <td className="py-6 text-[10px] font-black tracking-widest">{r.Method}</td>
                  <td className="py-6 text-xs text-zinc-400 font-semibold">{r.Service?.Name || 'auth-service'}</td>
                  <td className="py-6 text-[11px] font-bold text-emerald-500 flex items-center gap-2 mt-4.5">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" /> Active
                  </td>
                  <td className="py-6 text-right">
                    <div className="flex justify-end gap-2 opacity-20 group-hover:opacity-100 transition-opacity">
                      <button className="p-2 hover:bg-white/10 rounded-lg text-zinc-400 hover:text-white transition-all"><Edit3 size={16} /></button>
                      <button className="p-2 hover:bg-rose-500/10 rounded-lg text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={16} /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, trend, status, icon: Icon, color }: { label: string, value: string | number, trend: string, status: string, icon: any, color: string }) {
  const colorMap: any = {
    cyan: 'bg-cyan-500/20 text-cyan-400',
    purple: 'bg-purple-500/20 text-purple-400',
    emerald: 'bg-emerald-500/20 text-emerald-400',
    orange: 'bg-orange-500/20 text-orange-400'
  };
  return (
    <div className="bg-zinc-900/40 rounded-3xl border border-white/5 p-6 relative overflow-hidden group hover:border-cyan-500/20 transition-all">
      <div className="flex justify-between items-start mb-6">
        <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${colorMap[color]}`}><Icon size={20} /></div>
        <div className="flex flex-col items-end"><span className="text-[10px] text-zinc-500 font-black uppercase tracking-widest">{status}</span><span className="text-xs text-emerald-500 font-bold">{trend}</span></div>
      </div>
      <div className="flex flex-col"><span className="text-[11px] font-bold text-zinc-500 uppercase tracking-widest mb-1">{label}</span><span className="text-3xl font-black text-white">{value}</span></div>
    </div>
  );
}

function ServicesModule({ services, onRefresh }: { services: Service[], onRefresh: () => void }) {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center"><h3 className="text-xl font-bold">Manage Services</h3><button className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all"><Plus size={18} /><span>New Service</span></button></div>
      <div className="grid grid-cols-2 gap-6">
        {services.map(s => (
          <div key={s.ID} className="bg-zinc-900/40 rounded-3xl border border-white/5 p-6 flex items-center justify-between group hover:border-cyan-500/20 transition-all">
            <div className="flex items-center gap-5"><div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center text-cyan-400"><Server size={24} /></div><div className="flex flex-col"><span className="font-bold text-lg">{s.Name}</span><span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">{s.Protocol} • {s.Protocol === 'grpc' ? s.GRPCAddr : s.BaseURL}</span></div></div>
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"><button className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div>
          </div>
        ))}
      </div>
    </div>
  );
}

function RoutesModule({ routes, services, onRefresh }: { routes: Route[], services: Service[], onRefresh: () => void }) {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center"><h3 className="text-xl font-bold">Routing Control</h3><button className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all"><Plus size={18} /><span>Add Route</span></button></div>
      <div className="bg-zinc-900/40 rounded-3xl border border-white/5 overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-white/5"><tr><th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Route Path</th><th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Method</th><th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Upstream Target</th><th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Status</th><th className="p-6 text-right">Actions</th></tr></thead>
          <tbody className="divide-y divide-white/5">
            {routes.map(r => (
              <tr key={r.ID} className="group hover:bg-white/[0.02] transition-colors"><td className="p-6 font-mono text-sm text-cyan-400/80 group-hover:text-cyan-400 transition-colors">{r.Path}</td><td className="p-6 text-xs text-zinc-400 uppercase tracking-widest italic">{r.Method}</td><td className="p-6 text-sm font-semibold text-zinc-300">{r.Service?.Name || 'auth-service'}</td><td className="p-6 text-xs font-bold text-emerald-500">Active</td><td className="p-6 text-right"><div className="flex justify-end gap-3 opacity-0 group-hover:opacity-100 transition-opacity"><button className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div></td></tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ProtoMappingsModule({ mappings, onRefresh }: { mappings: ProtoMapping[], onRefresh: () => void }) {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">gRPC Proto Mappings</h3>
        <button className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all">
          <Plus size={18} />
          <span>New Mapping</span>
        </button>
      </div>
      <div className="bg-zinc-900/40 rounded-3xl border border-white/5 overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-white/5">
            <tr>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em]">RPC Method</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em]">Service</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em]">Package</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em]">Signature</th>
              <th className="p-6 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {mappings.map(m => (
              <tr key={m.ID} className="group hover:bg-white/[0.02] transition-colors">
                <td className="p-6 font-bold text-cyan-400/90">{m.RPCMethod}</td>
                <td className="p-6 text-sm font-semibold text-zinc-300">{m.Service?.Name || 'auth-service'}</td>
                <td className="p-6 font-mono text-xs text-zinc-500">{m.ProtoPackage}</td>
                <td className="p-6 text-[10px] font-mono"><div className="flex flex-col gap-0.5"><span className="text-cyan-500/50">Req: {m.RequestType}</span><span className="text-purple-500/50">Res: {m.ResponseType}</span></div></td>
                <td className="p-6 text-right"><div className="flex justify-end gap-3 opacity-0 group-hover:opacity-100 transition-opacity"><button className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
