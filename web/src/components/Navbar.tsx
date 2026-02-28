import { Link, useLocation } from 'react-router-dom';
import { Globe, LayoutDashboard, LogOut, Languages, ChevronRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '@/contexts/AuthContext';

export function Navbar() {
    const { t, i18n } = useTranslation();
    const { user, logout } = useAuth();
    const location = useLocation();

    const toggleLanguage = () => {
        i18n.changeLanguage(i18n.language.startsWith('zh') ? 'en' : 'zh');
    };

    const isActive = (path: string) =>
        location.pathname === path ? 'text-white' : 'text-zinc-400 hover:text-white';

    return (
        <header className="border-b border-white/[0.06] bg-[#09090b]/80 backdrop-blur-2xl sticky top-0 z-50">
            <div className="container mx-auto px-6 h-14 flex items-center justify-between max-w-6xl">
                <div className="flex items-center gap-8">
                    <Link to="/" className="flex items-center gap-2 group">
                        <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-blue-500/20">
                            <Globe className="w-4 h-4 text-white" />
                        </div>
                        <span className="font-bold text-white text-[15px] tracking-tight">Co-Link</span>
                    </Link>

                    <nav className="hidden md:flex items-center gap-1">
                        <Link to="/nodes" className={`text-sm px-3 py-1.5 rounded-md transition-colors flex items-center gap-1.5 ${isActive('/nodes')}`}>
                            <Globe className="w-3.5 h-3.5" /> {t('layout.nodes')}
                        </Link>
                        {user && (
                            <Link to="/dashboard" className={`text-sm px-3 py-1.5 rounded-md transition-colors flex items-center gap-1.5 ${isActive('/dashboard')}`}>
                                <LayoutDashboard className="w-3.5 h-3.5" /> {t('layout.dashboard')}
                            </Link>
                        )}
                    </nav>
                </div>

                <div className="flex items-center gap-2">
                    <button
                        onClick={toggleLanguage}
                        className="p-2 text-zinc-500 hover:text-zinc-300 transition-colors rounded-md hover:bg-white/5"
                    >
                        <Languages className="w-4 h-4" />
                    </button>

                    {user ? (
                        <div className="flex items-center gap-2">
                            <span className="text-xs text-zinc-500 hidden md:block">{user.email}</span>
                            <button
                                onClick={logout}
                                className="p-2 text-zinc-500 hover:text-red-400 transition-colors rounded-md hover:bg-red-500/5"
                                title={t('layout.logout')}
                            >
                                <LogOut className="w-4 h-4" />
                            </button>
                        </div>
                    ) : (
                        <div className="flex items-center gap-2">
                            <Link to="/login" className="text-sm text-zinc-400 hover:text-white px-3 py-1.5 rounded-md transition-colors">
                                {t('login.submit')}
                            </Link>
                            <Link
                                to="/register"
                                className="text-sm bg-white text-black hover:bg-zinc-100 px-3 py-1.5 rounded-md font-medium transition-colors flex items-center gap-1"
                            >
                                {t('register.submit')} <ChevronRight className="w-3.5 h-3.5" />
                            </Link>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
}
