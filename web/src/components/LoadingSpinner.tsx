export const LoadingSpinner = () => {
  return (
    <div className="flex justify-center items-center py-8">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2" style={{ borderColor: 'var(--color-primary)' }}></div>
      <span className="ml-3" style={{ color: 'var(--color-tertiary)' }}>Loading problems...</span>
    </div>
  )
}
