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
  Zap,
  Shield,
  RefreshCw,
  MoreVertical,
  X,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
  Filter,
  Terminal,
  Box
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
  RPCMethod: string;
  ServiceName: string;
  ProtoPackage: string;
  RequestType: string;
  ResponseType: string;
  Service?: Service;
}

interface ActivityLog {
  ID: number;
  Action: string;
  Resource: string;
  User: string;
  Message: string;
  CreatedAt: string;
}

interface RequestLog {
  ID: number;
  RequestID: string;
  Method: string;
  Path: string;
  StatusCode: number;
  LatencyMS: number;
  ClientIP: string;
  UserAgent: string;
  ErrorMessage: string;
  CreatedAt: string;
}

interface TraceLog {
  ID: number;
  RequestID: string;
  Level: string;
  Component: string;
  Message: string;
  CreatedAt: string;
}

interface ServerEntry {
  timestamp: string;
  message: string;
}

interface Metrics {
  services: Record<string, {
    total_requests: number;
    total_errors: number;
    avg_latency_ms: number;
    last_status: number;
    status_counts: Record<number, number>;
  }>;
}

const TABS = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'services', label: 'Services', icon: Server },
  { id: 'routes', label: 'Routes', icon: Globe },
  { id: 'proto', label: 'Proto Mappings', icon: Braces },
  { id: 'traffic', label: 'Traffic Monitor', icon: Activity },
  { id: 'system', label: 'System Logs', icon: Terminal },
  { id: 'settings', label: 'Settings', icon: Settings },
];

export default function App() {
  const [requestLogs, setRequestLogs] = useState<RequestLog[]>([]);
  const [activeTab, setActiveTab] = useState('dashboard');
  const [services, setServices] = useState<Service[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [protoMappings, setProtoMappings] = useState<ProtoMapping[]>([]);
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [modal, setModal] = useState<{ type: 'service' | 'route' | 'proto' | 'trace' | null, data: any }>({ type: null, data: null });
  const [serverLogs, setServerLogs] = useState<ServerEntry[]>([]);
  const [formLoading, setFormLoading] = useState(false);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchMetrics, 20000);
    return () => clearInterval(interval);
  }, [activeTab]);

  const fetchMetrics = async () => {
    try {
      const res = await axios.get('/admin/metrics');
      setMetrics(res.data);
    } catch (err) {
      console.error('Failed to fetch metrics', err);
    }
  };

  const fetchLogs = async () => {
    try {
      const res = await axios.get('/admin/logs');
      setLogs(res.data || []);
    } catch (err) {
      console.error('Failed to fetch logs', err);
    }
  };

  const fetchTrafficLogs = async () => {
    try {
      const res = await axios.get('/admin/request-logs');
      setRequestLogs(res.data || []);
    } catch (err) {
      console.error('Failed to fetch traffic logs', err);
    }
  };

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
      if (activeTab === 'dashboard') {
        await fetchMetrics();
        await fetchLogs();
      }
      if (activeTab === 'traffic' || activeTab === 'dashboard') {
        await fetchTrafficLogs();
      }
      if (activeTab === 'system') {
        const res = await axios.get('/admin/server-logs');
        setServerLogs(res.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch data', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (type: string, id: number) => {
    if (!confirm(`Are you sure you want to delete this ${type}?`)) return;
    try {
      await axios.delete(`/admin/${type === 'proto' ? 'proto-mappings' : type + 's'}/${id}`);
      fetchData();
    } catch (err) {
      console.error('Delete failed', err);
    }
  };

  const handleCreateOrUpdate = async (type: string, data: any) => {
    setFormLoading(true);
    try {
      const endpoint = `/admin/${type === 'proto' ? 'proto-mappings' : type + 's'}`;
      if (modal.data?.ID) {
        await axios.put(`${endpoint}/${modal.data.ID}`, data);
      } else {
        await axios.post(endpoint, data);
      }
      setModal({ type: null, data: null });
      fetchData();
    } catch (err) {
      console.error('Save failed', err);
    } finally {
      setFormLoading(false);
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
              {activeTab === 'dashboard' && <DashboardContent services={services} routes={routes} metrics={metrics} logs={logs} onRefresh={fetchData} setModal={setModal} handleDelete={handleDelete} />}
              {activeTab === 'services' && <ServicesModule services={services} onRefresh={fetchData} setModal={setModal} handleDelete={handleDelete} />}
              {activeTab === 'routes' && <RoutesModule routes={routes} services={services} onRefresh={fetchData} setModal={setModal} handleDelete={handleDelete} />}
              {activeTab === 'proto' && <ProtoMappingsModule mappings={protoMappings} onRefresh={fetchData} setModal={setModal} handleDelete={handleDelete} />}
              {activeTab === 'traffic' && <TrafficModule logs={requestLogs} onRefresh={fetchData} setModal={setModal} />}
              {activeTab === 'system' && <SystemLogsModule logs={serverLogs} onRefresh={fetchData} />}
            </motion.div>
          </AnimatePresence>
        </main>
      </div>

      <Modal 
        isOpen={!!modal.type} 
        onClose={() => setModal({ type: null, data: null })}
        title={modal.type === 'service' ? (modal.data?.ID ? 'Edit Service' : 'Add Service') : 
               modal.type === 'route' ? (modal.data?.ID ? 'Edit Route' : 'Add Route') : 
               modal.type === 'proto' ? (modal.data?.ID ? 'Edit Proto Mapping' : 'Add Proto Mapping') :
               modal.type === 'trace' ? 'Request Trace Timeline' : ''}
      >
        {modal.type === 'service' && <ServiceForm data={modal.data} onSubmit={(data) => handleCreateOrUpdate('service', data)} loading={formLoading} />}
        {modal.type === 'route' && <RouteForm data={modal.data} services={services} onSubmit={(data) => handleCreateOrUpdate('route', data)} loading={formLoading} />}
        {modal.type === 'proto' && <ProtoForm data={modal.data} services={services} onSubmit={(data) => handleCreateOrUpdate('proto', data)} loading={formLoading} />}
        {modal.type === 'trace' && <TraceView requestID={modal.data.RequestID} />}
      </Modal>
    </div>
  );
}

function DashboardContent({ services, routes, metrics, logs, onRefresh, setModal, handleDelete }: { services: Service[], routes: Route[], metrics: Metrics | null, logs: ActivityLog[], onRefresh: () => void, setModal: (val: any) => void, handleDelete: (type: string, id: number) => void }) {
  const totalRequests = metrics ? Object.values(metrics.services).reduce((acc, s) => acc + s.total_requests, 0) : 0;
  const avgLatency = metrics ? (Object.values(metrics.services).reduce((acc, s) => acc + s.avg_latency_ms, 0) / Object.keys(metrics.services).length || 0).toFixed(1) : '0';
  const onlineServices = services.filter(s => (s as any).Status === 'online').length;

  const [logSearch, setLogSearch] = useState('');
  const [logActionFilter, setLogActionFilter] = useState('all');
  const [logPage, setLogPage] = useState(1);
  const logsPerPage = 5;

  const filteredLogs = logs.filter(l => {
    const matchesSearch = l.Message.toLowerCase().includes(logSearch.toLowerCase()) || 
                          l.User.toLowerCase().includes(logSearch.toLowerCase());
    const matchesAction = logActionFilter === 'all' || l.Action === logActionFilter;
    return matchesSearch && matchesAction;
  });

  const totalLogPages = Math.ceil(filteredLogs.length / logsPerPage);
  const currentLogs = filteredLogs.slice((logPage - 1) * logsPerPage, logPage * logsPerPage);

  useEffect(() => setLogPage(1), [logSearch, logActionFilter]);

  return (
    <div className="grid grid-cols-12 gap-8">
      <div className="col-span-12 grid grid-cols-4 gap-6">
        <StatCard label="Active Services" value={`${onlineServices}/${services.length}`} trend="Real-time" status="Healthy" icon={Server} color="cyan" />
        <StatCard label="Total Requests" value={totalRequests} trend="+New" status="Live" icon={Globe} color="purple" />
        <StatCard label="Gateway Health" value="100%" trend="UP" status="Operational" icon={Shield} color="emerald" />
        <StatCard label="Avg. Latency" value={`${avgLatency}ms`} trend="Real-time" status="Stable" icon={Zap} color="orange" />
      </div>

      <div className="col-span-8 bg-zinc-900/40 rounded-3xl border border-white/5 p-8 relative overflow-hidden">
        <div className="flex justify-between items-center mb-10">
          <div>
            <h3 className="text-lg font-bold">Throughput per Service</h3>
            <p className="text-xs text-zinc-500 font-medium">Request distribution (Live metrics)</p>
          </div>
          <button onClick={onRefresh} className="p-2 bg-white/5 hover:bg-white/10 rounded-xl transition-all">
            <RefreshCw size={18} className="text-zinc-400" />
          </button>
        </div>
        <div className="h-[280px] flex items-end gap-6 px-4">
          {Object.entries(metrics?.services || {}).map(([name, data], i) => (
            <div key={i} className="flex-1 flex flex-col items-center gap-3">
               <div className="w-full bg-white/5 rounded-t-lg relative group transition-all" style={{ height: `${Math.min(data.total_requests * 5, 100)}%` }}>
                  <div className="absolute inset-0 bg-gradient-to-t from-cyan-600/20 to-cyan-400 scale-y-0 group-hover:scale-y-100 transition-transform origin-bottom rounded-t-lg" />
                  <div className="absolute -top-10 left-1/2 -translate-x-1/2 bg-cyan-500 text-black text-[10px] font-bold px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity">
                    {data.total_requests} req
                  </div>
               </div>
               <span className="text-[10px] text-zinc-500 font-bold uppercase truncate w-full text-center">{name}</span>
            </div>
          ))}
          {(!metrics || Object.keys(metrics.services).length === 0) && <div className="w-full text-center pb-20 text-zinc-600 font-medium">No activity data yet</div>}
        </div>
      </div>

      <div className="col-span-4 bg-zinc-900/40 rounded-3xl border border-white/5 p-8 flex flex-col">
        <div className="flex flex-col gap-4 mb-8">
          <h3 className="text-lg font-bold">System Activities</h3>
          <div className="space-y-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
              <input 
                type="text" 
                placeholder="Search logs..." 
                value={logSearch}
                onChange={(e) => setLogSearch(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-xl pl-9 pr-3 py-2 text-xs focus:outline-none focus:border-cyan-500/50"
              />
            </div>
            <div className="flex gap-2">
              {['all', 'CREATE', 'UPDATE', 'DELETE'].map(act => (
                <button 
                  key={act}
                  onClick={() => setLogActionFilter(act)}
                  className={`px-3 py-1 rounded-lg text-[10px] font-bold transition-all border ${logActionFilter === act ? 'bg-cyan-500/10 border-cyan-500/50 text-cyan-400' : 'bg-white/5 border-white/10 text-zinc-500 hover:text-zinc-400'}`}
                >
                  {act === 'all' ? 'ALL' : act}
                </button>
              ))}
            </div>
          </div>
        </div>
        <div className="space-y-6 flex-1 overflow-y-auto max-h-[350px] pr-2 custom-scrollbar">
          {currentLogs.map((log, i) => (
            <div key={log.ID} className="flex gap-4 relative">
              <div className={`w-2 h-2 rounded-full ${log.Action === 'DELETE' ? 'bg-rose-500' : log.Action === 'CREATE' ? 'bg-cyan-500' : 'bg-amber-500'} mt-1.5 shadow-[0_0_8px_rgba(255,255,255,0.2)]`} />
              <div className="flex flex-col gap-1">
                <span className="text-sm font-semibold text-zinc-200">{log.Message}</span>
                <span className="text-[10px] text-zinc-500 font-medium uppercase tracking-wider">{log.User} • {new Date(log.CreatedAt).toLocaleTimeString()}</span>
              </div>
            </div>
          ))}
          {currentLogs.length === 0 && <span className="text-zinc-600 text-sm italic">No matching activities...</span>}
        </div>
        {totalLogPages > 1 && (
          <div className="mt-8 flex justify-center gap-2">
            <button 
              disabled={logPage === 1} 
              onClick={() => setLogPage(logPage - 1)}
              className="p-1.5 bg-white/5 border border-white/10 rounded-lg text-zinc-500 disabled:opacity-20"
            >
              <ChevronLeft size={16} />
            </button>
            <span className="text-[10px] font-bold text-zinc-500 flex items-center">{logPage} / {totalLogPages}</span>
            <button 
              disabled={logPage === totalLogPages} 
              onClick={() => setLogPage(logPage + 1)}
              className="p-1.5 bg-white/5 border border-white/10 rounded-lg text-zinc-500 disabled:opacity-20"
            >
              <ChevronRight size={16} />
            </button>
          </div>
        )}
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
                      <button 
                        onClick={() => setModal({ type: 'route', data: r })}
                        className="p-2 hover:bg-white/10 rounded-lg text-zinc-400 hover:text-white transition-all"><Edit3 size={16} /></button>
                      <button 
                        onClick={() => handleDelete('route', r.ID)}
                        className="p-2 hover:bg-rose-500/10 rounded-lg text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={16} /></button>
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

function ServicesModule({ services, onRefresh, setModal, handleDelete }: { services: Service[], onRefresh: () => void, setModal: (v: any) => void, handleDelete: (t: string, id: number) => void }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [protocolFilter, setProtocolFilter] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 6;

  // Filtering logic
  const filteredItems = services.filter(s => {
    const matchesSearch = s.Name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                          (s.BaseURL || '').toLowerCase().includes(searchTerm.toLowerCase());
    const matchesProtocol = protocolFilter === 'all' || s.Protocol === protocolFilter;
    return matchesSearch && matchesProtocol;
  });

  const totalPages = Math.ceil(filteredItems.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const currentItems = filteredItems.slice(startIndex, startIndex + itemsPerPage);

  // Reset page on filter change
  useEffect(() => setCurrentPage(1), [searchTerm, protocolFilter]);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">Manage Services</h3>
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
            <input 
              type="text" 
              placeholder="Search services..." 
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-sm w-[200px] focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>
          <select 
            value={protocolFilter}
            onChange={(e) => setProtocolFilter(e.target.value)}
            className="bg-[#161618] border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:border-cyan-500/50"
          >
            <option value="all">All Protocols</option>
            <option value="rest">REST</option>
            <option value="grpc">gRPC</option>
          </select>
          <button onClick={() => setModal({ type: 'service' })} className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all"><Plus size={18} /><span>New Service</span></button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">
        {currentItems.map(s => (
          <div key={s.ID} className="bg-zinc-900/40 rounded-3xl border border-white/5 p-6 flex items-center justify-between group hover:border-cyan-500/20 transition-all">
            <div className="flex items-center gap-5"><div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center text-cyan-400"><Server size={24} /></div><div className="flex flex-col"><span className="font-bold text-lg">{s.Name}</span><span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">{s.Protocol} • {s.Protocol === 'grpc' ? s.GRPCAddr : s.BaseURL}</span></div></div>
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity"><button onClick={() => setModal({ type: 'service', data: s })} className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button onClick={() => handleDelete('service', s.ID)} className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div>
          </div>
        ))}
        {currentItems.length === 0 && <div className="col-span-2 text-center py-20 text-zinc-500 font-medium">No services match your criteria</div>}
      </div>
      <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
    </div>
  );
}

function RoutesModule({ routes, services, onRefresh, setModal, handleDelete }: { routes: Route[], services: Service[], onRefresh: () => void, setModal: (v: any) => void, handleDelete: (t: string, id: number) => void }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [methodFilter, setMethodFilter] = useState('all');
  const [sortConfig, setSortConfig] = useState<{ key: string, direction: 'asc' | 'desc' | null }>({ key: 'Path', direction: 'asc' });
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 8;

  // Filtering & Sorting Logic
  const filteredAndSortedItems = React.useMemo(() => {
    let result = routes.filter(r => {
      const matchesSearch = r.Path.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesMethod = methodFilter === 'all' || r.Method === methodFilter;
      return matchesSearch && matchesMethod;
    });

    if (sortConfig.key && sortConfig.direction) {
      result.sort((a: any, b: any) => {
        let aVal = a[sortConfig.key];
        let bVal = b[sortConfig.key];
        
        if (sortConfig.key === 'Service') {
          aVal = a.Service?.Name || '';
          bVal = b.Service?.Name || '';
        }

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }
    return result;
  }, [routes, searchTerm, methodFilter, sortConfig]);

  const totalPages = Math.ceil(filteredAndSortedItems.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const currentItems = filteredAndSortedItems.slice(startIndex, startIndex + itemsPerPage);

  useEffect(() => setCurrentPage(1), [searchTerm, methodFilter, sortConfig]);

  const handleSort = (key: string) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">Routing Control</h3>
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
            <input 
              type="text" 
              placeholder="Search paths..." 
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-sm w-[200px] focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>
          <select 
            value={methodFilter}
            onChange={(e) => setMethodFilter(e.target.value)}
            className="bg-[#161618] border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:border-cyan-500/50"
          >
            <option value="all">All Methods</option>
            {['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => <option key={m} value={m}>{m}</option>)}
          </select>
          <button onClick={() => setModal({ type: 'route' })} className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all"><Plus size={18} /><span>Add Route</span></button>
        </div>
      </div>

      <div className="bg-zinc-900/40 rounded-3xl border border-white/5 overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-white/5">
            <tr>
              {[
                { label: 'Route Path', key: 'Path' },
                { label: 'Method', key: 'Method' },
                { label: 'Upstream Target', key: 'Service' },
              ].map(col => (
                <th 
                  key={col.key} 
                  className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest cursor-pointer hover:text-zinc-300 transition-colors"
                  onClick={() => handleSort(col.key)}
                >
                  <div className="flex items-center gap-2">
                    {col.label}
                    <ArrowUpDown size={12} className={sortConfig.key === col.key ? 'text-cyan-400' : 'text-zinc-600'} />
                  </div>
                </th>
              ))}
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Status</th>
              <th className="p-6 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {currentItems.map(r => (
              <tr key={r.ID} className="group hover:bg-white/[0.02] transition-colors"><td className="p-6 font-mono text-sm text-cyan-400/80 group-hover:text-cyan-400 transition-colors">{r.Path}</td><td className="p-6 text-xs text-zinc-400 uppercase tracking-widest italic">{r.Method}</td><td className="p-6 text-sm font-semibold text-zinc-300">{r.Service?.Name || 'auth-service'}</td><td className="p-6 text-xs font-bold text-emerald-500">Active</td><td className="p-6 text-right"><div className="flex justify-end gap-3 opacity-0 group-hover:opacity-100 transition-opacity"><button onClick={() => setModal({ type: 'route', data: r })} className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button onClick={() => handleDelete('route', r.ID)} className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div></td></tr>
            ))}
            {currentItems.length === 0 && <tr><td colSpan={5} className="p-10 text-center text-zinc-500 font-medium">No routes found</td></tr>}
          </tbody>
        </table>
      </div>
      <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
    </div>
  );
}

function ProtoMappingsModule({ mappings, onRefresh, setModal, handleDelete }: { mappings: ProtoMapping[], onRefresh: () => void, setModal: (v: any) => void, handleDelete: (t: string, id: number) => void }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [packageFilter, setPackageFilter] = useState('all');
  const [sortConfig, setSortConfig] = useState<{ key: string, direction: 'asc' | 'desc' | null }>({ key: 'RPCMethod', direction: 'asc' });
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 8;

  const packages = Array.from(new Set(mappings.map(m => m.ProtoPackage)));

  const filteredAndSortedItems = React.useMemo(() => {
    let result = mappings.filter(m => {
      const matchesSearch = m.RPCMethod.toLowerCase().includes(searchTerm.toLowerCase()) || 
                            m.ProtoPackage.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesPackage = packageFilter === 'all' || m.ProtoPackage === packageFilter;
      return matchesSearch && matchesPackage;
    });

    if (sortConfig.key && sortConfig.direction) {
      result.sort((a: any, b: any) => {
        let aVal = a[sortConfig.key];
        let bVal = b[sortConfig.key];
        
        if (sortConfig.key === 'Service') {
          aVal = a.Service?.Name || '';
          bVal = b.Service?.Name || '';
        }

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }
    return result;
  }, [mappings, searchTerm, packageFilter, sortConfig]);

  const totalPages = Math.ceil(filteredAndSortedItems.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const currentItems = filteredAndSortedItems.slice(startIndex, startIndex + itemsPerPage);

  useEffect(() => setCurrentPage(1), [searchTerm, packageFilter, sortConfig]);

  const handleSort = (key: string) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">gRPC Proto Mappings</h3>
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
            <input 
              type="text" 
              placeholder="Search methods..." 
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-sm w-[200px] focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>
          <select 
            value={packageFilter}
            onChange={(e) => setPackageFilter(e.target.value)}
            className="bg-[#161618] border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:border-cyan-500/50"
          >
            <option value="all">All Packages</option>
            {packages.map(p => <option key={p} value={p}>{p}</option>)}
          </select>
          <button onClick={() => setModal({ type: 'proto' })} className="flex items-center gap-2 px-5 py-2.5 bg-cyan-500 hover:bg-cyan-600 rounded-xl text-sm font-bold transition-all"><Plus size={18} /><span>New Mapping</span></button>
        </div>
      </div>
      <div className="bg-zinc-900/40 rounded-3xl border border-white/5 overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-white/5">
            <tr>
              {[
                { label: 'RPC Method', key: 'RPCMethod' },
                { label: 'Service', key: 'Service' },
                { label: 'Package', key: 'ProtoPackage' }
              ].map(col => (
                <th key={col.key} className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em] cursor-pointer hover:text-zinc-300 transition-colors" onClick={() => handleSort(col.key)}>
                  <div className="flex items-center gap-2">
                    {col.label}
                    <ArrowUpDown size={12} className={sortConfig.key === col.key ? 'text-cyan-400' : 'text-zinc-600'} />
                  </div>
                </th>
              ))}
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-[0.2em]">Full Signature</th>
              <th className="p-6 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {currentItems.map(m => (
              <tr key={m.ID} className="group hover:bg-white/[0.02] transition-colors">
                <td className="p-6 font-bold text-cyan-400/90">{m.RPCMethod}</td>
                <td className="p-6 text-sm font-semibold text-zinc-300">{m.Service?.Name || 'auth-service'}</td>
                <td className="p-6 font-mono text-xs text-zinc-500">{m.ProtoPackage}.{m.ServiceName}</td>
                <td className="p-6 text-[10px] font-mono"><div className="flex flex-col gap-0.5"><span className="text-cyan-500/50">Method: {m.RPCMethod}</span><span className="text-purple-500/50">Types: {m.RequestType} → {m.ResponseType}</span></div></td>
                <td className="p-6 text-right"><div className="flex justify-end gap-3 opacity-0 group-hover:opacity-100 transition-opacity"><button onClick={() => setModal({ type: 'proto', data: m })} className="p-2 hover:bg-white/10 rounded-xl text-zinc-400 hover:text-white transition-all"><Edit3 size={18} /></button><button onClick={() => handleDelete('proto', m.ID)} className="p-2 hover:bg-rose-500/10 rounded-xl text-zinc-400 hover:text-rose-500 transition-all"><Trash2 size={18} /></button></div></td>
              </tr>
            ))}
            {currentItems.length === 0 && <tr><td colSpan={5} className="p-10 text-center text-zinc-500 font-medium">No mappings found</td></tr>}
          </tbody>
        </table>
      </div>
      <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
    </div>
  );
}

function TrafficModule({ logs, onRefresh, setModal }: { logs: RequestLog[], onRefresh: () => void, setModal: (m: any) => void }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [methodFilter, setMethodFilter] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const filteredLogs = React.useMemo(() => {
    return logs.filter(log => {
      const matchesSearch = log.Path.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            log.UserAgent.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesStatus = statusFilter === 'all' || 
                            (statusFilter === '200' && log.StatusCode < 300) ||
                            (statusFilter === '300' && log.StatusCode >= 300 && log.StatusCode < 400) ||
                            (statusFilter === '400' && log.StatusCode >= 400 && log.StatusCode < 500) ||
                            (statusFilter === '500' && log.StatusCode >= 500);
      const matchesMethod = methodFilter === 'all' || log.Method === methodFilter;
      return matchesSearch && matchesStatus && matchesMethod;
    });
  }, [logs, searchTerm, statusFilter, methodFilter]);

  const totalPages = Math.ceil(filteredLogs.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const currentItems = filteredLogs.slice(startIndex, startIndex + itemsPerPage);

  useEffect(() => setCurrentPage(1), [searchTerm, statusFilter, methodFilter]);

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'text-emerald-400';
    if (status >= 300 && status < 400) return 'text-blue-400';
    if (status >= 400 && status < 500) return 'text-amber-400';
    if (status >= 500) return 'text-rose-400';
    return 'text-zinc-400';
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">Live Traffic Logs</h3>
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
            <input 
              type="text" 
              placeholder="Search logs..." 
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-sm w-[200px] focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>
          <select 
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-[#161618] border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:border-cyan-500/50"
          >
            <option value="all">All Statuses</option>
            <option value="200">2xx Success</option>
            <option value="300">3xx Redirect</option>
            <option value="400">4xx Client Error</option>
            <option value="500">5xx Server Error</option>
          </select>
          <select 
            value={methodFilter}
            onChange={(e) => setMethodFilter(e.target.value)}
            className="bg-[#161618] border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:border-cyan-500/50"
          >
            <option value="all">All Methods</option>
            {['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => <option key={m} value={m}>{m}</option>)}
          </select>
          <button onClick={onRefresh} className="p-2 bg-white/5 hover:bg-white/10 rounded-xl transition-all">
            <RefreshCw size={18} className="text-zinc-400" />
          </button>
        </div>
      </div>

      <div className="bg-zinc-900/40 rounded-3xl border border-white/5 overflow-hidden">
        <table className="w-full text-left">
          <thead className="bg-white/5">
            <tr>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Timestamp</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Method</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Path</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Status</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Latency</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">Client IP</th>
              <th className="p-6 text-[11px] font-black text-zinc-500 uppercase tracking-widest">User Agent</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {currentItems.map(log => (
              <tr key={log.ID} className="hover:bg-white/[0.02] transition-colors cursor-pointer" onClick={() => setModal({ type: 'trace', data: log })}>
                <td className="p-6 text-xs text-zinc-400 tabular-nums">{new Date(log.CreatedAt).toLocaleTimeString()}</td>
                <td className="p-6 text-[10px] font-black tracking-widest">{log.Method}</td>
                <td className="p-6 font-mono text-xs text-cyan-400/80">{log.Path}</td>
                <td className={`p-6 text-xs font-bold ${getStatusColor(log.StatusCode)}`}>{log.StatusCode}</td>
                <td className="p-6 text-xs text-zinc-400 tabular-nums">{log.LatencyMS}ms</td>
                <td className="p-6 text-xs text-zinc-400 font-mono">{log.ClientIP}</td>
                <td className="p-6 text-xs text-zinc-400 truncate max-w-[150px]" title={log.UserAgent}>{log.UserAgent}</td>
              </tr>
            ))}
            {currentItems.length === 0 && <tr><td colSpan={7} className="p-20 text-center text-zinc-600 italic font-medium">No traffic logs found</td></tr>}
          </tbody>
        </table>
      </div>
      <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
    </div>
  );
}


function TraceView({ requestID }: { requestID: string }) {
  const [traces, setTraces] = useState<TraceLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
      const tryFetch = async () => {
        try {
          const res = await axios.get(`/admin/traces/${requestID}`);
          setTraces(Array.isArray(res.data) ? res.data : []);
        } catch (err) {
          console.error('Failed to fetch traces', err);
          setTraces([]);
        } finally {
          setLoading(false);
        }
      };
      tryFetch();
  }, [requestID]);

  if (loading) return <div className="p-20 flex justify-center"><RefreshCw className="animate-spin text-cyan-500" /></div>;

  return (
    <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-4 custom-scrollbar">
      {traces.length === 0 && <div className="text-center py-10 text-zinc-500 italic">No trace events recorded for this request.</div>}
      {traces.map((t) => (
        <div key={t.ID} className="relative pl-8 pb-4 border-l border-white/10 last:pb-0">
          <div className={`absolute left-[-5px] top-1.5 w-2 h-2 rounded-full ${t.Level === 'ERROR' ? 'bg-rose-500 shadow-[0_0_10px_rgba(244,63,94,0.5)]' : t.Level === 'WARN' ? 'bg-amber-500' : 'bg-cyan-500'}`} />
          <div className="bg-white/5 rounded-xl p-4 border border-white/5 hover:border-white/10 transition-colors">
            <div className="flex justify-between items-center mb-1">
              <span className={`text-[10px] font-black tracking-widest uppercase ${t.Level === 'ERROR' ? 'text-rose-500' : t.Level === 'WARN' ? 'text-amber-500' : 'text-cyan-500'}`}>{t.Level} • {t.Component}</span>
              <span className="text-[10px] text-zinc-600 font-mono tabular-nums">{new Date(t.CreatedAt).toLocaleTimeString()}</span>
            </div>
            <p className="text-sm text-zinc-300 leading-relaxed font-mono">{t.Message}</p>
          </div>
        </div>
      ))}
    </div>
  );
}

function SystemLogsModule({ logs, onRefresh }: { logs: ServerEntry[], onRefresh: () => void }) {
  const logsEndRef = React.useRef<HTMLDivElement>(null);

  useEffect(() => {
    const interval = setInterval(onRefresh, 5000);
    return () => clearInterval(interval);
  }, [onRefresh]);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h3 className="text-xl font-bold">System Console</h3>
        <button onClick={onRefresh} className="p-2 bg-white/5 hover:bg-white/10 rounded-xl transition-all">
          <RefreshCw size={18} className="text-zinc-400" />
        </button>
      </div>
      <div className="bg-[#0c0c0e] rounded-3xl border border-white/5 p-6 font-mono text-xs overflow-hidden h-[70vh] flex flex-col">
        <div className="flex-1 overflow-y-auto space-y-1 custom-scrollbar pr-4">
          {logs.map((l, i) => (
            <div key={i} className="flex gap-4 group">
              <span className="text-zinc-700 whitespace-nowrap tabular-nums">{new Date(l.timestamp).toLocaleTimeString()}</span>
              <span className="text-zinc-400 break-all">{l.message}</span>
            </div>
          ))}
          {logs.length === 0 && <div className="text-center py-20 text-zinc-700 italic">Listening for server output...</div>}
          <div ref={logsEndRef} />
        </div>
      </div>
    </div>
  );
}

function Modal({ isOpen, onClose, title, children }: { isOpen: boolean, onClose: () => void, title: string, children: React.ReactNode }) {
  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <motion.div 
            initial={{ opacity: 0 }} 
            animate={{ opacity: 1 }} 
            exit={{ opacity: 0 }} 
            onClick={onClose}
            className="absolute inset-0 bg-black/60 backdrop-blur-sm" 
          />
          <motion.div 
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 20 }}
            className="relative w-full max-w-lg bg-[#0d0d0f] border border-white/10 rounded-3xl shadow-2xl overflow-hidden"
          >
            <div className="p-6 border-b border-white/5 flex justify-between items-center">
              <h3 className="text-lg font-bold">{title}</h3>
              <button onClick={onClose} className="p-2 hover:bg-white/5 rounded-xl transition-all text-zinc-400 hover:text-zinc-200">
                <X size={20} />
              </button>
            </div>
            <div className="p-6">
              {children}
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}

function ServiceForm({ data, onSubmit, loading }: { data: any, onSubmit: (val: any) => void, loading: boolean }) {
  const [formData, setFormData] = React.useState({
    Name: data?.Name || '',
    BaseURL: data?.BaseURL || '',
    Protocol: data?.Protocol || 'rest',
    GRPCAddr: data?.GRPCAddr || '',
  });

  return (
    <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); onSubmit(formData); }}>
      <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Service Name</label><input value={formData.Name} onChange={(e) => setFormData({ ...formData, Name: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="e.g. auth-service" required /></div>
      <div>
        <label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Protocol</label>
        <select value={formData.Protocol} onChange={(e) => setFormData({ ...formData, Protocol: e.target.value })} className="w-full bg-[#161618] border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50">
          <option value="rest">REST</option>
          <option value="grpc">gRPC</option>
        </select>
      </div>
      {formData.Protocol === 'rest' ? (
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Base URL</label><input value={formData.BaseURL} onChange={(e) => setFormData({ ...formData, BaseURL: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="http://localhost:8081" required /></div>
      ) : (
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">gRPC Address</label><input value={formData.GRPCAddr} onChange={(e) => setFormData({ ...formData, GRPCAddr: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="localhost:50051" required /></div>
      )}
      <button disabled={loading} className="w-full py-3 bg-cyan-500 hover:bg-cyan-600 rounded-xl font-bold text-sm transition-all mt-6 shadow-[0_0_20px_rgba(6,182,212,0.3)] disabled:opacity-50">
        {loading ? 'Saving...' : 'Save Service'}
      </button>
    </form>
  );
}

function RouteForm({ data, services, onSubmit, loading }: { data: any, services: Service[], onSubmit: (val: any) => void, loading: boolean }) {
  const [formData, setFormData] = React.useState({
    Path: data?.Path || '',
    Method: data?.Method || 'GET',
    ServiceID: data?.ServiceID || services[0]?.ID || 0,
    EndpointFilter: data?.EndpointFilter || '',
    Tag: data?.Tag || 'default',
  });

  return (
    <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); onSubmit({ ...formData, ServiceID: Number(formData.ServiceID) }); }}>
      <div className="flex gap-4">
        <div className="w-1/3">
          <label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Method</label>
          <select value={formData.Method} onChange={(e) => setFormData({ ...formData, Method: e.target.value })} className="w-full bg-[#161618] border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50">
            {['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>
        <div className="flex-1">
          <label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Path</label>
          <input value={formData.Path} onChange={(e) => setFormData({ ...formData, Path: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="/v1/auth/login" required />
        </div>
      </div>
      <div>
        <label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Upstream Service</label>
        <select value={formData.ServiceID} onChange={(e) => setFormData({ ...formData, ServiceID: Number(e.target.value) })} className="w-full bg-[#161618] border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50">
          {services.map(s => <option key={s.ID} value={s.ID}>{s.Name}</option>)}
        </select>
      </div>
      <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Endpoint Filter / Handler</label><input value={formData.EndpointFilter} onChange={(e) => setFormData({ ...formData, EndpointFilter: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="login" required /></div>
      <button disabled={loading} className="w-full py-3 bg-cyan-500 hover:bg-cyan-600 rounded-xl font-bold text-sm transition-all mt-6 shadow-[0_0_20px_rgba(6,182,212,0.3)] disabled:opacity-50">
        {loading ? 'Saving...' : 'Save Route'}
      </button>
    </form>
  );
}

function ProtoForm({ data, services, onSubmit, loading }: { data: any, services: Service[], onSubmit: (val: any) => void, loading: boolean }) {
  const [formData, setFormData] = React.useState({
    ServiceID: data?.ServiceID || services.find(s => s.Protocol === 'grpc')?.ID || 0,
    RPCMethod: data?.RPCMethod || '',
    ServiceName: data?.ServiceName || '',
    ProtoPackage: data?.ProtoPackage || '',
    RequestType: data?.RequestType || '',
    ResponseType: data?.ResponseType || '',
  });

  return (
    <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); onSubmit({ ...formData, ServiceID: Number(formData.ServiceID) }); }}>
      <div>
        <label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">GRPC Service</label>
        <select value={formData.ServiceID} onChange={(e) => setFormData({ ...formData, ServiceID: Number(e.target.value) })} className="w-full bg-[#161618] border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50">
          {services.filter(s => s.Protocol === 'grpc').map(s => <option key={s.ID} value={s.ID}>{s.Name}</option>)}
        </select>
      </div>
      <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">RPC Method Name</label><input value={formData.RPCMethod} onChange={(e) => setFormData({ ...formData, RPCMethod: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="e.g. Login" required /></div>
      <div className="grid grid-cols-2 gap-4">
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Proto Package</label><input value={formData.ProtoPackage} onChange={(e) => setFormData({ ...formData, ProtoPackage: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="e.g. auth.v1" required /></div>
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">gRPC Service Name</label><input value={formData.ServiceName} onChange={(e) => setFormData({ ...formData, ServiceName: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="e.g. AuthService" required /></div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Request Type</label><input value={formData.RequestType} onChange={(e) => setFormData({ ...formData, RequestType: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="LoginRequest" required /></div>
        <div><label className="text-[10px] font-black uppercase tracking-widest text-zinc-500 mb-2 block">Response Type</label><input value={formData.ResponseType} onChange={(e) => setFormData({ ...formData, ResponseType: e.target.value })} className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-cyan-500/50" placeholder="LoginResponse" required /></div>
      </div>
      <button disabled={loading} className="w-full py-3 bg-cyan-500 hover:bg-cyan-600 rounded-xl font-bold text-sm transition-all mt-6 shadow-[0_0_20px_rgba(6,182,212,0.3)] disabled:opacity-50">
        {loading ? 'Saving...' : 'Save Mapping'}
      </button>
    </form>
  );
}

function Pagination({ currentPage, totalPages, onPageChange }: { currentPage: number, totalPages: number, onPageChange: (page: number) => void }) {
  if (totalPages <= 1) return null;
  return (
    <div className="flex items-center justify-center gap-2 mt-8">
      <button 
        disabled={currentPage === 1}
        onClick={() => onPageChange(currentPage - 1)}
        className="p-2 bg-white/5 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/10 transition-all disabled:opacity-20"
      >
        <ChevronLeft size={18} />
      </button>
      <div className="flex items-center gap-1">
        {Array.from({ length: totalPages }).map((_, i) => (
          <button 
            key={i}
            onClick={() => onPageChange(i + 1)}
            className={`w-10 h-10 rounded-xl font-bold text-xs transition-all ${currentPage === i+1 ? 'bg-cyan-500 text-black shadow-[0_0_15px_rgba(6,182,212,0.4)]' : 'bg-white/5 border border-white/10 text-zinc-500 hover:text-white hover:bg-white/10'}`}
          >
            {i + 1}
          </button>
        ))}
      </div>
      <button 
        disabled={currentPage === totalPages}
        onClick={() => onPageChange(currentPage + 1)}
        className="p-2 bg-white/5 border border-white/10 rounded-xl text-zinc-400 hover:text-white hover:bg-white/10 transition-all disabled:opacity-20"
      >
        <ChevronRight size={18} />
      </button>
    </div>
  );
}
