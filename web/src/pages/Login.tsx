import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { api } from '@/lib/api';
import { Loader2, ArrowRight, Languages, Globe } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export default function Login() {
    const { t, i18n } = useTranslation();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const toggleLanguage = () => {
        i18n.changeLanguage(i18n.language.startsWith('zh') ? 'en' : 'zh');
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');
        try {
            const res = await api.post('/auth/login', { email, password });
            login(res.data.token, res.data.user);
            navigate('/dashboard');
        } catch (err: any) {
            setError(err.response?.data?.error || t('login.error'));
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex flex-col">
            {/* Top bar */}
            <div className="flex items-center justify-between px-6 py-4">
                <Link to="/" className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-blue-500/20">
                        <Globe className="w-4 h-4 text-white" />
                    </div>
                    <span className="font-bold text-white text-[15px]">Co-Link</span>
                </Link>
                <button onClick={toggleLanguage} className="p-2 text-zinc-600 hover:text-zinc-300 transition-colors rounded-md hover:bg-white/5">
                    <Languages className="w-4 h-4" />
                </button>
            </div>

            {/* Center card */}
            <div className="flex-1 flex items-center justify-center px-6 py-12">
                <div className="w-full max-w-sm">
                    <div className="mb-8">
                        <h1 className="text-2xl font-bold text-white mb-1.5">{t('login.title')}</h1>
                        <p className="text-zinc-500 text-sm">{t('login.subtitle')}</p>
                    </div>

                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <div className="px-4 py-3 rounded-lg bg-red-500/8 border border-red-500/20 text-red-400 text-sm">
                                {error}
                            </div>
                        )}

                        <div className="space-y-1.5">
                            <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{t('login.email')}</label>
                            <input
                                type="email"
                                value={email}
                                onChange={e => setEmail(e.target.value)}
                                required
                                placeholder="you@example.com"
                                className="w-full px-3.5 py-2.5 rounded-lg bg-white/[0.04] border border-white/[0.08] text-white placeholder:text-zinc-700 focus:outline-none focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all text-sm"
                            />
                        </div>

                        <div className="space-y-1.5">
                            <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{t('login.password')}</label>
                            <input
                                type="password"
                                value={password}
                                onChange={e => setPassword(e.target.value)}
                                required
                                placeholder="••••••••"
                                className="w-full px-3.5 py-2.5 rounded-lg bg-white/[0.04] border border-white/[0.08] text-white placeholder:text-zinc-700 focus:outline-none focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 transition-all text-sm"
                            />
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-4 py-2.5 rounded-lg text-sm font-semibold transition-all flex items-center justify-center gap-2 group shadow-lg shadow-blue-600/20 mt-2"
                        >
                            {isLoading ? <Loader2 className="animate-spin w-4 h-4" /> : (
                                <>{t('login.submit')} <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" /></>
                            )}
                        </button>
                    </form>

                    <p className="mt-6 text-center text-sm text-zinc-600">
                        {t('login.noAccount')}{' '}
                        <Link to="/register" className="text-blue-400 hover:text-blue-300 transition-colors font-medium">
                            {t('login.register')}
                        </Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
