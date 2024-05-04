import { Line } from 'react-chartjs-2';
import { Chart as ChartJS, ArcElement, Tooltip, Legend, CategoryScale, LinearScale, PointElement, LineElement } from 'chart.js';

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, PointElement, LineElement)

export default function TestLine() {
    return (
        <Line
          data={{
            labels: ['Jun', 'Jul', 'Aug'],
            datasets: [
              {
                id: 1,
                label: '',
                data: [5, 6, 7],
              },
              {
                id: 2,
                label: '',
                data: [3, 2, 1],
              },
            ],
          }}
        />
    );
}