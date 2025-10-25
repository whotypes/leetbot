interface Problem {
  id: number
  url: string
  title: string
  difficulty: string
  acceptance: number
  frequency: number
}

interface ProblemsTableProps {
  problems: Problem[]
}

const getDifficultyColor = (difficulty: string) => {
  switch (difficulty.toLowerCase()) {
    case 'easy':
      return { backgroundColor: '#dcfce7', color: '#166534' }
    case 'medium':
      return { backgroundColor: '#fef3c7', color: '#d97706' }
    case 'hard':
      return { backgroundColor: '#fee2e2', color: '#dc2626' }
    default:
      return { backgroundColor: 'var(--color-muted)', color: 'var(--color-content)' }
  }
}

export const ProblemsTable = ({ problems }: ProblemsTableProps) => {
  return (
    <div className="rounded-lg border overflow-hidden" style={{ backgroundColor: 'var(--color-surface)', borderColor: 'var(--color-muted)' }}>
      <div className="px-6 py-4 border-b" style={{ borderColor: 'var(--color-muted)' }}>
        <h2 className="text-lg font-semibold" style={{ color: 'var(--color-content)' }}>
          Problems ({problems.length})
        </h2>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full" style={{ borderCollapse: 'separate', borderSpacing: 0 }}>
          <thead style={{ backgroundColor: 'var(--color-surface)' }}>
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--color-tertiary)' }}>
                ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--color-tertiary)' }}>
                Title
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--color-tertiary)' }}>
                Difficulty
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--color-tertiary)' }}>
                Acceptance
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider" style={{ color: 'var(--color-tertiary)' }}>
                Frequency
              </th>
            </tr>
          </thead>
          <tbody style={{ backgroundColor: 'var(--color-surface)' }}>
            {problems.map((problem) => (
              <tr key={problem.id} className="hover:opacity-80" style={{ borderBottom: '1px solid var(--color-muted)' }}>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium" style={{ color: 'var(--color-content)' }}>
                  {problem.id}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <a
                    href={problem.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm hover:underline"
                    style={{ color: 'var(--color-primary)' }}
                  >
                    {problem.title}
                  </a>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full" style={getDifficultyColor(problem.difficulty)}>
                    {problem.difficulty}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm" style={{ color: 'var(--color-content)' }}>
                  {problem.acceptance.toFixed(1)}%
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm" style={{ color: 'var(--color-content)' }}>
                  {problem.frequency.toFixed(1)}%
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
