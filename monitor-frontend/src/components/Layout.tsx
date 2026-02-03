import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { LayoutDashboard, Activity, Layers, FileCode, Hexagon } from 'lucide-react';
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
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <LayoutDashboard size={20} />
                        <span>Dashboard</span>
                    </NavLink>
                    <NavLink 
                        to="/liquidations" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <Activity size={20} />
                        <span>Liquidations</span>
                    </NavLink>
                    <NavLink 
                        to="/transfers" 
                        className={({ isActive }) => `${styles.navItem} ${isActive ? styles.active : ''}`}
                    >
                        <Layers size={20} />
                        <span>Transfers</span>
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
                    <div className={styles.version}>v3.0.0</div>
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
