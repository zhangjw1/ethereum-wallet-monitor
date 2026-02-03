import React from 'react';
import styles from './KPICard.module.css';
import { TrendingUp, TrendingDown } from 'lucide-react';

interface KPICardProps {
    title: string;
    value: string;
    change?: string;
    isPositive?: boolean;
}

const KPICard: React.FC<KPICardProps> = ({ title, value, change, isPositive }) => {
    return (
        <div className={styles.card}>
            <h3 className={styles.title}>{title}</h3>
            <div className={styles.value}>{value}</div>
            {change && (
                <div className={`${styles.change} ${isPositive ? styles.positive : styles.negative}`}>
                    {isPositive ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                    <span>{change}</span>
                </div>
            )}
            <div className={styles.sparkline}>
                {/* Simplified sparkline */}
                <svg viewBox="0 0 100 20" className={styles.chart}>
                    <path
                        d="M0 15 Q 25 5, 50 10 T 100 5"
                        fill="none"
                        stroke={isPositive ? "var(--success)" : "var(--danger)"}
                        strokeWidth="2"
                    />
                </svg>
            </div>
        </div>
    );
};

export default KPICard;
