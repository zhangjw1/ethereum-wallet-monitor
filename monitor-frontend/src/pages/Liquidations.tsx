import React from 'react';
import styles from './Liquidations.module.css';

const Liquidations: React.FC = () => {
    const liquidations = [
        { id: 1, time: '2 mins ago', asset: 'WETH', amount: '$124,500', profit: '+$4,200', tx: '0x8a...9c1' },
        { id: 2, time: '5 mins ago', asset: 'WBTC', amount: '$450,000', profit: '+$12,050', tx: '0x3f...e2a' },
        { id: 3, time: '12 mins ago', asset: 'USDC', amount: '$85,000', profit: '+$1,100', tx: '0x7b...d4f' },
        { id: 4, time: '15 mins ago', asset: 'DAI', amount: '$200,000', profit: '+$2,500', tx: '0x1c...a5b' },
        { id: 5, time: '22 mins ago', asset: 'AAVE', amount: '$45,000', profit: '+$890', tx: '0x9e...f3d' },
    ];

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <h2>Liquidation Radar</h2>
                <div className={styles.filters}>
                    <select className={styles.select}>
                        <option>All Assets</option>
                        <option>WETH</option>
                        <option>WBTC</option>
                    </select>
                </div>
            </div>

            <div className={`card ${styles.tableCard}`}>
                <table className={styles.table}>
                    <thead>
                        <tr>
                            <th>Time</th>
                            <th>Asset</th>
                            <th>Liquidated Amt</th>
                            <th>Est. Profit</th>
                            <th>Tx Hash</th>
                            <th>Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {liquidations.map((liq) => (
                            <tr key={liq.id}>
                                <td className={styles.time}>{liq.time}</td>
                                <td className={styles.asset}>{liq.asset}</td>
                                <td>{liq.amount}</td>
                                <td>
                                    <span className="badge badge-success">{liq.profit}</span>
                                </td>
                                <td className={styles.hash}>{liq.tx}</td>
                                <td>
                                    <button className={`${styles.actionBtn} btn-outline`}>View</button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default Liquidations;
