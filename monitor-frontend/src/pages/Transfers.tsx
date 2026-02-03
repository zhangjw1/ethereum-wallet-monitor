import React from 'react';
import styles from './Transfers.module.css';
import { ArrowUpRight, ArrowDownLeft } from 'lucide-react';

const Transfers: React.FC = () => {
    const transfers = [
        { id: 1, type: 'inbound', from: 'Binance Hot Wallet', to: 'Whale 0x8a...', amount: '500 ETH', value: '$1.2M', time: 'Just now' },
        { id: 2, type: 'outbound', from: 'Whale 0x3f...', to: 'Uniswap V3', amount: '1,000,000 USDC', value: '$1.0M', time: '45s ago' },
        { id: 3, type: 'inbound', from: 'Unknown', to: 'Vitalik.eth', amount: '100 ETH', value: '$240k', time: '2m ago' },
        { id: 4, type: 'outbound', from: 'Whale 0x9e...', to: 'Aave V3', amount: '50 WBTC', value: '$2.8M', time: '5m ago' },
        { id: 5, type: 'inbound', from: 'Coinbase', to: 'Whale 0x1c...', amount: '250,000 LINK', value: '$3.5M', time: '12m ago' },
    ];

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <div className={styles.pageInfo}>
                    <h2>Whale Watch</h2>
                    <p>Tracking high-value transfers across major protocols.</p>
                </div>
                <div className={styles.stats}>
                    <div className={styles.statItem}>
                        <span className={styles.statLabel}>24h Volume</span>
                        <span className={styles.statValue}>$450M</span>
                    </div>
                </div>
            </div>

            <div className={styles.feed}>
                {transfers.map((tx) => (
                    <div key={tx.id} className={styles.logItem}>
                        <div className={`${styles.iconBase} ${tx.type === 'inbound' ? styles.iconIn : styles.iconOut}`}>
                            {tx.type === 'inbound' ? <ArrowDownLeft size={20} /> : <ArrowUpRight size={20} />}
                        </div>
                        <div className={styles.content}>
                            <div className={styles.rowTop}>
                                <span className={styles.addresses}>
                                    <span className={styles.addr}>{tx.from}</span>
                                    <span className={styles.arrow}>→</span>
                                    <span className={styles.addr}>{tx.to}</span>
                                </span>
                                <span className={styles.amount}>{tx.amount}</span>
                            </div>
                            <div className={styles.rowBottom}>
                                <span className={styles.time}>{tx.time}</span>
                                <span className={styles.value}>{tx.value}</span>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default Transfers;
