interface CompanySelectorProps {
  companies: string[]
  selectedCompany: string
  onCompanyChange: (company: string) => void
}

export const CompanySelector = ({ companies, selectedCompany, onCompanyChange }: CompanySelectorProps) => {
  return (
    <div>
      <label htmlFor="company-select" className="block text-sm font-medium mb-2" style={{ color: 'var(--color-content)' }}>
        Company
      </label>
      <select
        id="company-select"
        value={selectedCompany}
        onChange={(e) => onCompanyChange(e.target.value)}
        className="w-full p-4 border rounded-md focus:outline-none focus:ring-2"
        style={{
          borderColor: 'var(--color-muted)',
          backgroundColor: 'var(--color-surface)',
          color: 'var(--color-content)',
          '--tw-ring-color': 'var(--color-primary)'
        } as React.CSSProperties & { '--tw-ring-color': string }}
      >
        <option value="">Select a company</option>
        {companies.map((company) => (
          <option key={company} value={company}>
            {company.charAt(0).toUpperCase() + company.slice(1).replace(/-/g, ' ')}
          </option>
        ))}
      </select>
    </div>
  )
}
