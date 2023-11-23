import React, { useEffect } from 'react';

export function CategoryList({transactions}) {
    let [categoryData, setCategoryData] = React.useState(null);
    useEffect(() => {
        const data = fillOutData(transactions)
        console.log(data)
        setCategoryData(data)}, [transactions]
    );
    
    return categoryData && <ul> 
     {makeItemList(categoryData)}
    </ul>
  }

  function makeItemList(dataMap) {
    let list = []
    dataMap.forEach((val, key, map) => {
        list.push(<li>{key}: {val}</li>);
    })
    return list;
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
      return keyVal;
  }


  