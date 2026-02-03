import React, { useEffect, useState } from 'react';
import KPICard from '../components/KPICard';
import CodePreview from '../components/CodePreview';
import styles from './Dashboard.module.css';

const Dashboard: React.FC = () => {
    // Mock data for live stream
    const [streamData, setStreamData] = useState<any>({
        event: "Transfer",
        token: "PEPE",
        amount: "1,500,000,000",
        from: "0x3f...e2a",
        to: "0x8a...9c1",
        timestamp: new Date().toISOString()
    });

    useEffect(() => {
        const interval = setInterval(() => {
            const tokens = ["PEPE", "WETH", "USDT", "LINK"];
            const amounts = ["1,000", "50.5", "10,000", "500"];

            setStreamData({
                event: Math.random() > 0.5 ? "Transfer" : "Swap",
                token: tokens[Math.floor(Math.random() * tokens.length)],
                amount: amounts[Math.floor(Math.random() * amounts.length)],
                from: "0x" + Math.random().toString(16).substr(2, 6) + "...",
                to: "0x" + Math.random().toString(16).substr(2, 6) + "...",
                timestamp: new Date().toISOString(),
                txHash: "0x" + Math.random().toString(16).substr(2, 40)
            });
        }, 3000);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className={styles.container}>
            <section className={styles.hero}>
                <div className={styles.heroContent}>
                    <h2 className={styles.heroTitle}>System Operational</h2>
                    <p className={styles.heroSubtitle}>Monitoring 185 contracts across Ethereum Mainnet.</p>
                </div>
                <div className={styles.heroActions}>
                    <button className="btn btn-primary">Connect Wallet</button>
                    <button className="btn btn-outline">View Logs</button>
                </div>
            </section>

            <section className={styles.stats}>
                <KPICard
                    title="Total Value Locked"
                    value="$142.5M"
                    change="+12.5%"
                    isPositive={true}
                />
                <KPICard
                    title="24h Volume"
                    value="$38.2M"
                    change="+5.2%"
                    isPositive={true}
                />
                <KPICard
                    title="Active Alerts"
                    value="12"
                    change="-2"
                    isPositive={true}
                />
                <KPICard
                    title="Network Gas"
                    value="24 Gwei"
                    change="+15%"
                    isPositive={false}
                />
            </section>

            <section className={styles.integration}>
                <div className={styles.integrationInfo}>
                    <h3 className={styles.sectionTitle}>Real-time API Stream</h3>
                    <p className={styles.sectionText}>
                        Integrate directly with our WebSocket API to get sub-second updates on liquidations and whale movements.
                    </p>
                    <div className={styles.apiCodes}>
                        <code>wss://api.ethermonitor.io/v1/stream</code>
                    </div>
                </div>
                <div className={styles.previewContainer}>
                    <CodePreview title="WebSocket Feed (Active)" data={streamData} />
                </div>
            </section>
        </div>
    );
};

export default Dashboard;
