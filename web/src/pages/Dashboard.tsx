import React, { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { Copy, Check, Terminal, FileJson, KeySquare, Shield, Activity, Cpu } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Navbar } from '@/components/Navbar';

export default function Dashboard() {
    const { user } = useAuth();
    const { t } = useTranslation();
    const [copiedAPI, setCopiedAPI] = useState(false);
    const [copiedClient, setCopiedClient] = useState(false);

    if (!user) return null;

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const httpProtocol = window.location.protocol;

    const copyToClipboard = (text: string, setter: React.Dispatch<React.SetStateAction<boolean>>) => {
        navigator.clipboard.writeText(text);
        setter(true);
        setTimeout(() => setter(false), 2000);
    };

    return (
        <div className="min-h-screen">
            <Navbar />

            <div className="max-w-4xl mx-auto px-6 py-10">
                {/* Header */}
                <div className="mb-10">
                    <p className="text-xs text-zinc-600 uppercase tracking-widest mb-2">
                        {user.email}
                    </p>
                    <h1 className="text-3xl font-bold text-white">{t('dashboard.title')}</h1>
                    <p className="text-zinc-500 text-sm mt-1.5">{t('dashboard.subtitle')}</p>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-2 gap-4 mb-6">
                    <div className="rounded-2xl border border-white/[0.06] bg-gradient-to-br from-white/[0.04] to-transparent p-5">
                        <div className="flex items-center gap-2 mb-2">
                            <Activity className="w-4 h-4 text-emerald-400" />
                            <span className="text-zinc-400 text-sm font-medium">{t('dashboard.apiCalls')}</span>
                        </div>
                        <div className="text-3xl font-light text-white tracking-tight">
                            {user.total_api_calls || 0}
                        </div>
                    </div>
                    <div className="rounded-2xl border border-white/[0.06] bg-gradient-to-br from-white/[0.04] to-transparent p-5">
                        <div className="flex items-center gap-2 mb-2">
                            <Cpu className="w-4 h-4 text-amber-400" />
                            <span className="text-zinc-400 text-sm font-medium">{t('dashboard.providedCalls')}</span>
                        </div>
                        <div className="text-3xl font-light text-white tracking-tight">
                            {user.total_provided_calls || 0}
                        </div>
                    </div>
                </div>

                {/* Token Cards */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                    {/* API Token */}
                    <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5 group hover:bg-white/[0.04] hover:border-white/10 transition-all">
                        <div className="flex items-center gap-3 mb-4">
                            <div className="p-2 rounded-lg bg-blue-500/10">
                                <KeySquare className="w-4 h-4 text-blue-400" />
                            </div>
                            <div>
                                <h3 className="text-sm font-semibold text-white">{t('dashboard.apiToken')}</h3>
                                <p className="text-xs text-zinc-600 mt-0.5">{t('dashboard.apiTokenDesc')}</p>
                            </div>
                        </div>
                        <div className="flex items-center gap-2 p-2.5 rounded-lg bg-black/40 border border-white/[0.06] font-mono text-xs">
                            <span className="truncate flex-1 text-zinc-400 select-all">{user.api_token}</span>
                            <button
                                onClick={() => copyToClipboard(user.api_token, setCopiedAPI)}
                                className="flex-shrink-0 p-1.5 hover:bg-white/10 rounded-md transition-colors text-zinc-600 hover:text-white"
                            >
                                {copiedAPI ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                            </button>
                        </div>
                    </div>

                    {/* Client Token */}
                    <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5 group hover:bg-white/[0.04] hover:border-white/10 transition-all">
                        <div className="flex items-center gap-3 mb-4">
                            <div className="p-2 rounded-lg bg-indigo-500/10">
                                <Shield className="w-4 h-4 text-indigo-400" />
                            </div>
                            <div>
                                <h3 className="text-sm font-semibold text-white">{t('dashboard.clientToken')}</h3>
                                <p className="text-xs text-zinc-600 mt-0.5">{t('dashboard.clientTokenDesc')}</p>
                            </div>
                        </div>
                        <div className="flex items-center gap-2 p-2.5 rounded-lg bg-black/40 border border-white/[0.06] font-mono text-xs">
                            <span className="truncate flex-1 text-zinc-400 select-all">{user.client_token}</span>
                            <button
                                onClick={() => copyToClipboard(user.client_token, setCopiedClient)}
                                className="flex-shrink-0 p-1.5 hover:bg-white/10 rounded-md transition-colors text-zinc-600 hover:text-white"
                            >
                                {copiedClient ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Client Config */}
                <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden mb-4">
                    <div className="px-5 py-3.5 border-b border-white/[0.06] flex items-center gap-2">
                        <FileJson className="w-4 h-4 text-blue-400" />
                        <span className="text-sm font-medium text-white">{t('dashboard.clientConfig')}</span>
                        <span className="ml-auto text-xs text-zinc-600">{t('dashboard.clientConfigDesc')}</span>
                    </div>
                    <div className="p-5 bg-black/40">
                        <pre className="font-mono text-xs leading-6 text-zinc-400 overflow-x-auto">
                            <p><span className="text-blue-400">client_token:</span> <span className="text-green-300">"{user.client_token}"</span></p>
                            <p><span className="text-blue-400">server_url:</span> <span className="text-green-300">"{wsProtocol}//{host}/ws"</span></p>
                            <p><span className="text-blue-400">max_parallel:</span> <span className="text-orange-300">3</span></p>
                            <p><span className="text-blue-400">providers:</span></p>
                            <p>{'  '}- <span className="text-purple-400">type:</span> <span className="text-green-300">"openai"</span></p>
                            <p>{'    '}<span className="text-purple-400">api_key:</span> <span className="text-green-300">"sk-your-real-api-key"</span></p>
                            <p>{'    '}<span className="text-purple-400">base_url:</span> <span className="text-green-300">"{httpProtocol}//{host}/v1"</span></p>
                            <p>{'    '}<span className="text-purple-400">models:</span></p>
                            <p>{'      '}- <span className="text-cyan-400">local:</span> <span className="text-green-300">"gpt-4-turbo"</span></p>
                            <p>{'        '}<span className="text-cyan-400">server_mapping:</span> <span className="text-green-300">"pro-model"</span></p>
                        </pre>
                    </div>
                </div>

                {/* API Test */}
                <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden">
                    <div className="px-5 py-3.5 border-b border-white/[0.06] flex items-center gap-2">
                        <Terminal className="w-4 h-4 text-indigo-400" />
                        <span className="text-sm font-medium text-white">{t('dashboard.apiTest')}</span>
                        <span className="ml-auto text-xs text-zinc-600">{t('dashboard.apiTestDesc')}</span>
                    </div>
                    <div className="p-5 bg-black/40">
                        <pre className="font-mono text-xs leading-6 text-zinc-400 overflow-x-auto">
                            <p>
                                <span className="text-blue-400">curl</span>
                                {' '}-X POST <span className="text-green-300">{httpProtocol}//{host}/v1/chat/completions</span> \
                            </p>
                            <p>
                                {'  '}-H <span className="text-yellow-300">"Authorization: Bearer <span className="text-green-300">{user.api_token}</span>"</span> \
                            </p>
                            <p>
                                {'  '}-H <span className="text-yellow-300">"Content-Type: application/json"</span> \
                            </p>
                            <p>
                                {'  '}-d <span className="text-zinc-500">'{"{"}  "model": "pro-model", "stream": true, "messages": [{"{"}"role": "user", "content": "Hello!"{"}"}] {"}"}"</span>
                            </p>
                        </pre>
                    </div>
                </div>
            </div>
        </div>
    );
}
