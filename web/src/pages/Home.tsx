import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowRight, Globe, Shield, Zap, Server, Code, Terminal, ChevronRight } from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { Navbar } from '@/components/Navbar';

export default function Home() {
    const { t } = useTranslation();
    const { user } = useAuth();

    return (
        <div className="min-h-screen">
            <Navbar />

            {/* Hero */}
            <section className="relative pt-28 pb-24 px-6 text-center overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-b from-blue-500/[0.07] via-transparent to-transparent pointer-events-none" />

                <div className="relative z-10 max-w-3xl mx-auto">
                    <div className="inline-flex items-center gap-2 mb-8 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-300 text-xs font-medium">
                        <span className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
                        {t('home.badge')}
                    </div>

                    <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-6 leading-[1.05]">
                        <span className="text-white">{t('home.heroTitle').split('.')[0]}.</span>
                        {t('home.heroTitle').includes('.') && (
                            <>
                                <br />
                                <span className="bg-gradient-to-r from-blue-400 via-indigo-400 to-violet-400 bg-clip-text text-transparent">
                                    {t('home.heroTitle').split('.').slice(1).join('.').trim()}
                                </span>
                            </>
                        )}
                    </h1>

                    <p className="text-lg text-zinc-500 max-w-xl mx-auto mb-10 leading-relaxed">
                        {t('home.heroSubtitle')}
                    </p>

                    <div className="flex items-center justify-center gap-3 flex-wrap">
                        {user ? (
                            <Link to="/dashboard" className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-5 py-2.5 rounded-xl text-sm font-semibold transition-all shadow-lg shadow-blue-500/20 group">
                                {t('layout.dashboard')} <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                            </Link>
                        ) : (
                            <Link to="/register" className="inline-flex items-center gap-2 bg-white text-black hover:bg-zinc-100 px-5 py-2.5 rounded-xl text-sm font-semibold transition-all group">
                                {t('home.ctaStart')} <ChevronRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                            </Link>
                        )}
                        <Link to="/nodes" className="inline-flex items-center gap-2 bg-white/5 hover:bg-white/10 border border-white/[0.08] text-zinc-300 hover:text-white px-5 py-2.5 rounded-xl text-sm font-medium transition-all">
                            <Globe className="w-4 h-4" /> {t('home.ctaNodes')}
                        </Link>
                    </div>
                </div>

                <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[400px] bg-blue-600/8 rounded-full blur-[100px] pointer-events-none" />
            </section>

            {/* Stats bar */}
            <section className="border-y border-white/[0.06] py-6 px-6">
                <div className="max-w-5xl mx-auto flex flex-wrap items-center justify-center gap-8 md:gap-16">
                    {[
                        { label: 'OpenAI Compatible', value: '100%' },
                        { label: 'Protocol', value: 'WebSocket' },
                        { label: 'Auth', value: 'JWT + bcrypt' },
                        { label: 'Streaming', value: 'SSE Native' },
                    ].map(s => (
                        <div key={s.label} className="text-center">
                            <div className="text-xl font-bold text-white">{s.value}</div>
                            <div className="text-xs text-zinc-600 mt-0.5">{s.label}</div>
                        </div>
                    ))}
                </div>
            </section>

            {/* Features */}
            <section className="py-24 px-6">
                <div className="max-w-5xl mx-auto">
                    <div className="text-center mb-16">
                        <h2 className="text-3xl font-bold text-white mb-3">{t('home.featuresTitle')}</h2>
                        <p className="text-zinc-500">{t('home.featuresSubtitle')}</p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {[
                            { icon: Zap, color: 'text-yellow-400', bg: 'bg-yellow-500/10', border: 'border-yellow-500/20', titleKey: 'home.feat1Title', descKey: 'home.feat1Desc' },
                            { icon: Shield, color: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/20', titleKey: 'home.feat2Title', descKey: 'home.feat2Desc' },
                            { icon: Server, color: 'text-indigo-400', bg: 'bg-indigo-500/10', border: 'border-indigo-500/20', titleKey: 'home.feat3Title', descKey: 'home.feat3Desc' },
                        ].map(({ icon: Icon, color, bg, border, titleKey, descKey }) => (
                            <div key={titleKey} className="group rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 hover:bg-white/[0.04] hover:border-white/10 transition-all duration-300">
                                <div className={`p-2.5 mb-5 rounded-xl w-fit border ${bg} ${border} ${color}`}>
                                    <Icon className="w-5 h-5" />
                                </div>
                                <h3 className="font-semibold text-white mb-2">{t(titleKey)}</h3>
                                <p className="text-zinc-500 text-sm leading-relaxed">{t(descKey)}</p>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* Quick Start */}
            <section className="py-24 px-6 border-t border-white/[0.06]">
                <div className="max-w-5xl mx-auto">
                    <div className="text-center mb-14">
                        <h2 className="text-3xl font-bold text-white mb-3">{t('home.quickStartTitle')}</h2>
                        <p className="text-zinc-500">{t('home.quickStartSubtitle')}</p>
                    </div>

                    {/* Step 1 — Register (full width) */}
                    <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden mb-4">
                        <div className="px-5 py-4 flex items-center gap-3 border-b border-white/[0.06]">
                            <span className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-500/20 text-blue-300 text-xs flex items-center justify-center font-bold">1</span>
                            <Shield className="w-4 h-4 text-zinc-500" />
                            <span className="text-sm font-medium text-white">{t('home.step2Title')}</span>
                        </div>
                        <div className="p-5 flex flex-col sm:flex-row sm:items-center gap-4">
                            <p className="text-zinc-500 text-sm flex-1">{t('home.step2Desc')}</p>
                            <div className="flex gap-2 flex-shrink-0">
                                <Link to="/register" className="inline-flex items-center gap-1.5 bg-white text-black hover:bg-zinc-100 px-4 py-2 rounded-lg text-sm font-medium transition-colors">
                                    {t('register.submit')} <ArrowRight className="w-3.5 h-3.5" />
                                </Link>
                                <Link to="/login" className="inline-flex items-center gap-1.5 border border-white/10 hover:border-white/20 text-zinc-400 hover:text-white px-4 py-2 rounded-lg text-sm transition-colors">
                                    {t('login.submit')}
                                </Link>
                            </div>
                        </div>
                    </div>

                    {/* Steps 2 & 3 — Side by side */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">

                        {/* Left — API Usage */}
                        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden flex flex-col">
                            <div className="px-5 py-3.5 flex items-center gap-3 border-b border-white/[0.06]">
                                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 text-xs flex items-center justify-center font-bold">2</span>
                                <Terminal className="w-4 h-4 text-zinc-500" />
                                <span className="text-sm font-medium text-white">API 接入</span>
                                <span className="ml-auto text-[10px] px-2 py-0.5 rounded-md bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 font-mono">OpenAI Compatible</span>
                            </div>
                            <div className="p-5 flex flex-col gap-3 flex-1">
                                <p className="text-zinc-500 text-sm">
                                    将您的 <code className="text-zinc-300 bg-white/5 px-1 rounded">API Token</code> 作为 Bearer 密钥，替换 base_url 后所有 OpenAI SDK 均可直接使用。
                                </p>
                                <div className="bg-black/60 rounded-lg p-4 font-mono text-[11px] text-zinc-400 border border-white/[0.06] leading-[1.8] overflow-x-auto flex-1">
                                    <p className="text-zinc-600 mb-1"># Python SDK</p>
                                    <p><span className="text-blue-400">from</span> openai <span className="text-blue-400">import</span> OpenAI</p>
                                    <p>client = OpenAI(</p>
                                    <p>{'  '}api_key=<span className="text-yellow-300">"sk-colink-your-token"</span>,</p>
                                    <p>{'  '}base_url=<span className="text-yellow-300">"https://your-server/v1"</span>,</p>
                                    <p>)</p>
                                    <p className="mt-2 text-zinc-600"># curl</p>
                                    <p><span className="text-green-400">curl</span> -X POST https://your-server/v1/chat/completions \</p>
                                    <p>{'  '}-H <span className="text-yellow-300">"Authorization: Bearer sk-colink-..."</span> \</p>
                                    <p>{'  '}-d <span className="text-yellow-300">'{"{"}model":"pro-model","stream":true,...{"}"}'</span></p>
                                </div>
                            </div>
                        </div>

                        {/* Right — Client Config */}
                        <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] overflow-hidden flex flex-col">
                            <div className="px-5 py-3.5 flex items-center gap-3 border-b border-white/[0.06]">
                                <span className="flex-shrink-0 w-6 h-6 rounded-full bg-violet-500/20 text-violet-300 text-xs flex items-center justify-center font-bold">3</span>
                                <Code className="w-4 h-4 text-zinc-500" />
                                <span className="text-sm font-medium text-white">客户端接入</span>
                                <span className="ml-auto text-[10px] px-2 py-0.5 rounded-md bg-violet-500/10 border border-violet-500/20 text-violet-300 font-mono">config.yaml</span>
                            </div>
                            <div className="p-5 flex flex-col gap-3 flex-1">
                                <p className="text-zinc-500 text-sm">
                                    将您的 <code className="text-zinc-300 bg-white/5 px-1 rounded">Client Token</code> 写入配置文件，运行 client 进程即可将本地 GPU 加入网络。
                                </p>
                                <div className="bg-black/60 rounded-lg p-4 font-mono text-[11px] text-zinc-400 border border-white/[0.06] leading-[1.8] overflow-x-auto flex-1">
                                    <p><span className="text-blue-400">client_token:</span> <span className="text-green-300">"client-your-token"</span></p>
                                    <p><span className="text-blue-400">server_url:</span> <span className="text-green-300">"wss://your-server/ws"</span></p>
                                    <p><span className="text-blue-400">max_parallel:</span> <span className="text-orange-300">3</span></p>
                                    <p><span className="text-blue-400">providers:</span></p>
                                    <p>{'  '}- <span className="text-purple-400">type:</span> <span className="text-green-300">"openai"</span></p>
                                    <p>{'    '}<span className="text-purple-400">api_key:</span> <span className="text-green-300">"sk-your-real-key"</span></p>
                                    <p>{'    '}<span className="text-purple-400">base_url:</span> <span className="text-green-300">"https://api.openai.com/v1"</span></p>
                                    <p>{'    '}<span className="text-purple-400">models:</span></p>
                                    <p>{'      '}- <span className="text-cyan-400">local:</span> <span className="text-green-300">"gpt-4-turbo"</span></p>
                                    <p>{'        '}<span className="text-cyan-400">server_mapping:</span> <span className="text-green-300">"pro-model"</span></p>
                                    <p className="mt-2 text-zinc-600"># 启动客户端</p>
                                    <p><span className="text-green-400">./client</span> -config config.yaml</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* CTA Footer */}
            {!user && (
                <section className="py-20 px-6 border-t border-white/[0.06]">
                    <div className="max-w-lg mx-auto text-center">
                        <h2 className="text-2xl font-bold text-white mb-3">{t('home.ctaStart')}</h2>
                        <p className="text-zinc-500 text-sm mb-8">{t('home.heroSubtitle')}</p>
                        <Link to="/register" className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-6 py-3 rounded-xl text-sm font-semibold transition-all shadow-lg shadow-blue-500/20 group">
                            {t('home.ctaStart')} <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
                        </Link>
                    </div>
                </section>
            )}

            <footer className="border-t border-white/[0.06] py-6 text-center">
                <span className="text-xs text-zinc-700">Co-Link Plan · {new Date().getFullYear()}</span>
            </footer>
        </div>
    );
}
