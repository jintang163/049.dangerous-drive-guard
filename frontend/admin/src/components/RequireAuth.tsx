import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { getToken, isTokenExpired } from '@/utils/auth'

interface Props {
  children: React.ReactNode
}

const RequireAuth: React.FC<Props> = ({ children }) => {
  const token = getToken()
  const location = useLocation()

  if (!token || isTokenExpired()) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }

  return <>{children}</>
}

export default RequireAuth
