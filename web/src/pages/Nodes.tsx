import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { Server, Activity, Cpu, Layers } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Navbar } from '@/components/Navbar';

interface NodeInfo {
    id: string;
    max_parallel: number;
    active_tasks: number;
    supported_models: string[];
    penalized: boolean;
}

export default function Nodes() {
    const { t } = useTranslation();
    const [nodes, setNodes] = useState<NodeInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

    useEffect(() => {
        const fetchNodes = async () => {
            try {
                const res = await api.get('/nodes');
                setNodes(res.data.nodes || []);
                setLastUpdated(new Date());
            } catch (e) {
                console.error('Failed to fetch nodes', e);
            } finally {
                setLoading(false);
            }
        };

        fetchNodes();
        const interval = setInterval(fetchNodes, 5000);
        return () => clearInterval(interval);
    }, []);

    const totalCapacity = nodes.reduce((sum, n) => sum + n.max_parallel, 0);
    const totalActive = nodes.reduce((sum, n) => sum + n.active_tasks, 0);
    const healthyCount = nodes.filter(n => !n.penalized).length;

    return (
        <div className="min-h-screen">
            <Navbar />

            <div className="max-w-6xl mx-auto px-6 py-10">
                {/* Header */}
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-10">
                    <div>
                        <h1 className="text-3xl font-bold text-white mb-1.5">{t('nodes.title')}</h1>
                        <p className="text-zinc-500 text-sm">{t('nodes.subtitle')}</p>
                    </div>

                    <div className="flex items-center gap-2 self-start">
                        <div className="flex items-center gap-2 bg-white/[0.03] border border-white/[0.06] px-3 py-1.5 rounded-full">
                            <span className="relative flex h-2 w-2">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-60" />
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
                            </span>
                            <span className="text-xs font-medium text-zinc-300">{nodes.length} {t('nodes.online')}</span>
                        </div>
                        {lastUpdated && (
                            <span className="text-xs text-zinc-600 hidden sm:block">
                                {lastUpdated.toLocaleTimeString()}
                            </span>
                        )}
                    </div>
                </div>

                {/* Summary stats */}
                {nodes.length > 0 && (
                    <div className="grid grid-cols-3 gap-4 mb-8">
                        {[
                            { icon: Server, label: t('nodes.online'), value: `${healthyCount} / ${nodes.length}`, color: 'text-green-400' },
                            { icon: Activity, label: t('nodes.capacity'), value: `${totalActive} / ${totalCapacity}`, color: 'text-blue-400' },
                            { icon: Cpu, label: 'Utilization', value: totalCapacity > 0 ? `${Math.round((totalActive / totalCapacity) * 100)}%` : '0%', color: 'text-indigo-400' },
                        ].map(s => (
                            <div key={s.label} className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4 flex items-center gap-3">
                                <s.icon className={`w-5 h-5 ${s.color} flex-shrink-0`} />
                                <div>
                                    <div className="text-lg font-bold text-white">{s.value}</div>
                                    <div className="text-xs text-zinc-600">{s.label}</div>
                                </div>
                            </div>
                        ))}
                    </div>
                )}

                {/* Node grid */}
                {loading ? (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {[1, 2, 3].map(i => (
                            <div key={i} className="h-44 rounded-2xl bg-white/[0.02] border border-white/[0.06] animate-pulse" />
                        ))}
                    </div>
                ) : nodes.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-24 rounded-2xl border border-dashed border-white/10">
                        <Server className="w-10 h-10 text-zinc-700 mb-4" />
                        <h3 className="text-base font-medium text-zinc-400 mb-1">{t('nodes.emptyTitle')}</h3>
                        <p className="text-sm text-zinc-600">{t('nodes.emptyDesc')}</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {nodes.map((node, i) => {
                            const utilization = node.max_parallel > 0 ? Math.round((node.active_tasks / node.max_parallel) * 100) : 0;
                            const barColor = node.penalized ? 'bg-orange-500' : utilization > 80 ? 'bg-red-500' : utilization > 50 ? 'bg-yellow-400' : 'bg-blue-500';

                            return (
                                <div
                                    key={node.id || i}
                                    className={`rounded-2xl border bg-white/[0.02] p-5 relative overflow-hidden transition-all duration-300 hover:-translate-y-0.5 hover:bg-white/[0.04] ${node.penalized ? 'border-orange-500/30' : 'border-white/[0.06] hover:border-white/10'
                                        }`}
                                >
                                    {/* Top accent */}
                                    <div className={`absolute top-0 inset-x-0 h-px ${node.penalized ? 'bg-gradient-to-r from-orange-500/50 via-red-500/50 to-transparent' : 'bg-gradient-to-r from-blue-500/20 via-transparent to-transparent'}`} />

                                    <div className="flex items-start justify-between mb-5">
                                        <div className="flex items-center gap-3">
                                            <div className={`p-2 rounded-lg ${node.penalized ? 'bg-orange-500/10' : 'bg-blue-500/10'}`}>
                                                <Server className={`w-4 h-4 ${node.penalized ? 'text-orange-400' : 'text-blue-400'}`} />
                                            </div>
                                            <div>
                                                <p className="font-mono text-xs text-zinc-300 leading-none mb-1">
                                                    {node.id ? `...${node.id.split('-').pop()}` : `node-${i}`}
                                                </p>
                                                <div className="flex items-center gap-1.5">
                                                    <Activity className="w-3 h-3 text-zinc-600" />
                                                    <span className={`text-[11px] font-medium ${node.penalized ? 'text-orange-400' : 'text-green-400'}`}>
                                                        {node.penalized ? t('nodes.penalized') : t('nodes.healthy')}
                                                    </span>
                                                </div>
                                            </div>
                                        </div>
                                        <span className={`text-[10px] font-mono px-2 py-1 rounded-md ${node.penalized ? 'bg-orange-500/10 text-orange-400' : 'bg-green-500/10 text-green-400'}`}>
                                            {node.active_tasks}/{node.max_parallel}
                                        </span>
                                    </div>

                                    {/* Utilization bar */}
                                    <div className="mb-4">
                                        <div className="flex justify-between mb-1.5">
                                            <span className="text-[10px] text-zinc-600 uppercase tracking-wider">{t('nodes.capacity')}</span>
                                            <span className="text-[10px] text-zinc-500">{utilization}%</span>
                                        </div>
                                        <div className="h-1 w-full bg-white/5 rounded-full overflow-hidden">
                                            <div
                                                className={`h-full rounded-full transition-all duration-700 ${barColor}`}
                                                style={{ width: `${Math.max(utilization, utilization > 0 ? 3 : 0)}%` }}
                                            />
                                        </div>
                                    </div>

                                    {/* Models */}
                                    {node.supported_models.length > 0 && (
                                        <div>
                                            <div className="flex items-center gap-1.5 mb-2">
                                                <Layers className="w-3 h-3 text-zinc-600" />
                                                <span className="text-[10px] text-zinc-600 uppercase tracking-wider">{t('nodes.models')}</span>
                                            </div>
                                            <div className="flex flex-wrap gap-1.5">
                                                {node.supported_models.map(m => (
                                                    <span key={m} className="px-2 py-0.5 bg-blue-500/8 border border-blue-500/15 text-blue-300 text-[10px] rounded font-mono">
                                                        {m}
                                                    </span>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
