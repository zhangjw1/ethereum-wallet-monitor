import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { LayoutDashboard, Layers, Shield, Bell, FileCode, Hexagon } from 'lucide-react';
import styles from './Layout.module.css';

const Layout: React.FC = () => {
    return (
        <div className={styles.layout}>
            <aside className={styles.sidebar}>
                <div className={styles.logo}>
                    <Hexagon className={styles.logoIcon} size={28} strokeWidth={2.5} />
                    <span className={styles.logoText}>EtherMonitor</span>
                </div>
                
                <nav className={styles.nav}>
                    <NavLink 
                        to="/" 
                        end
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <LayoutDashboard size={20} />
                        <span>Dashboard</span>
                    </NavLink>
                    <NavLink 
                        to="/transfers" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <Layers size={20} />
                        <span>交易流水</span>
                    </NavLink>
                    <NavLink 
                        to="/tokens" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <Shield size={20} />
                        <span>代币分析</span>
                    </NavLink>
                    <NavLink 
                        to="/notifications" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <Bell size={20} />
                        <span>通知记录</span>
                    </NavLink>
                    <NavLink 
                        to="/api-docs" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <FileCode size={20} />
                        <span>API Docs</span>
                    </NavLink>
                </nav>

                <div className={styles.footer}>
                    <div className={styles.status}>
                        <div className={styles.statusDot}></div>
                        <span>System Operational</span>
                    </div>
                    <div className={styles.version}>v3.1.0</div>
                </div>
            </aside>
            <main className={styles.main}>
                <header className={styles.header}>
                    <h1 className={styles.pageTitle}>Overview</h1>
                    <div className={styles.userControls}>
                        <div className={styles.avatar}>A</div>
                    </div>
                </header>
                <div className={styles.content}>
                    <Outlet />
                </div>
            </main>
        </div>
    );
};

export default Layout;
