import { useState } from "react"

export default function BudgetUpdateForm(props: { handleBudgetUpdate: (newBudget: number) => void }) {
    let [budget, setBudget] = useState<string>('0')
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        let numBudget = parseFloat(budget)
        const response = await fetch("http://localhost:8080/api/budget/set", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify({ budget: numBudget }),
        });
        props.handleBudgetUpdate(numBudget)
    }
    return (
        < form onSubmit={handleSubmit} >
            <label>
                Set new budget:
                <input value={budget} onChange={(e) => setBudget(e.target.value)} />
            </label>
            <input type="submit" />
        </form >
    )
}
