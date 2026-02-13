import { ReactNode } from 'react'

interface CardProps {
  children: ReactNode
  className?: string
  id?: string
}

export default function Card({ children, className = '', id }: CardProps) {
  return (
    <div id={id} className={`bg-dark-bg-elevated border border-dark-border-subtle rounded-xl shadow-linear ${className}`}>
      {children}
    </div>
  )
}

export function CardHeader({ children, className = '' }: CardProps) {
  return (
    <div className={`px-8 pt-8 pb-4 ${className}`}>
      {children}
    </div>
  )
}

export function CardBody({ children, className = '' }: CardProps) {
  return (
    <div className={`px-8 pb-8 ${className}`}>
      {children}
    </div>
  )
}
