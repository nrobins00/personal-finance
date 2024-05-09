import { Pie } from 'react-chartjs-2';
import { Chart as ChartJS, ArcElement, Colors } from "chart.js";

ChartJS.register(ArcElement);


export default function BudgetSpendingPie(props: { budget: number, spending: number }) {
    return (
        <div style={{ height: '100px', width: '100px' }}>
            <Pie
                datasetIdKey='id'
                data={{
                    datasets: [
                        {
                            label: 'budget',
                            data: [props.budget + props.spending, props.spending],
                            backgroundColor: [
                                'rgba(255, 99, 132, 0.2)',
                                'rgba(54, 162, 235, 0.2)',
                            ]
                        },
                    ]
                }}
            />
        </div>
    )
}
