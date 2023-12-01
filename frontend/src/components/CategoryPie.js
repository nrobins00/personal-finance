import React, { useEffect } from 'react';
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js';
import { Doughnut } from 'react-chartjs-2';

ChartJS.register(ArcElement, Tooltip, Legend);

export function CategoryPie({transactions}) {
    let [pieData, setPieData] = React.useState(null);
    useEffect(() => {
        const data = fillOutData(transactions)
        console.log(data)
        setPieData(data)}, [transactions]
    );
    
    return pieData && <> 
     <div>
     <Doughnut data={pieData} options={{
        width: 100,
        height: 100,
        
      }} />
     </div>
    </>
  }

  function fillOutData(transactions) {
    if (transactions.length < 1) return null;
    const keyVal = new Map()
    console.log(keyVal.size);
    transactions.forEach(element => {
        let category = element.category[0];
        if (keyVal.has(category)) {
            let currentSum = keyVal.get(element.category[0]);
            console.log("category: " + category + "current sum: " + currentSum)
            keyVal.set(category, currentSum + element.amount)
        }
        else {
            console.log(category, element.amount)
            keyVal.set(category, element.amount)
        }
    });
    const data = {
        labels: Array.from(keyVal.keys()),
        
        datasets: [
          {
            label: '$',
            data: Array.from(keyVal.values()),
            backgroundColor: [
                'rgba(255, 99, 132, 0.2)',
                'rgba(54, 162, 235, 0.2)',
                'rgba(255, 206, 86, 0.2)',
                'rgba(75, 192, 192, 0.2)',
                'rgba(153, 102, 255, 0.2)',
                'rgba(255, 159, 64, 0.2)',
              ],
            borderWidth: 1,
          },
        ],
      };
      console.log(data);
      return data;
  }


  /*
  backgroundColor: [
              'rgba(255, 99, 132, 0.2)',
              'rgba(54, 162, 235, 0.2)',
              'rgba(255, 206, 86, 0.2)',
              'rgba(75, 192, 192, 0.2)',
              'rgba(153, 102, 255, 0.2)',
              'rgba(255, 159, 64, 0.2)',
            ],
            borderColor: [
              'rgba(255, 99, 132, 1)',
              'rgba(54, 162, 235, 1)',
              'rgba(255, 206, 86, 1)',
              'rgba(75, 192, 192, 1)',
              'rgba(153, 102, 255, 1)',
              'rgba(255, 159, 64, 1)',
            ],
*/
  
