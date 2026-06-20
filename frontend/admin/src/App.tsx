import React from 'react'
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import MainLayout from '@/layouts/MainLayout'
import DashboardLayout from '@/layouts/DashboardLayout'
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import Monitor from '@/pages/Monitor'
import RoutePlan from '@/pages/RoutePlan'
import FatigueAlarms from '@/pages/FatigueAlarms'
import Vehicles from '@/pages/Vehicles'
import Drivers from '@/pages/Drivers'
import Waybills from '@/pages/Waybills'
import Escort from '@/pages/Escort'
import Rescue from '@/pages/Rescue'
import Weather from '@/pages/Weather'
import Blockchain from '@/pages/Blockchain'
import Profile from '@/pages/Profile'
import NotFound from '@/pages/NotFound'
import RequireAuth from '@/components/RequireAuth'
import { useAppStore } from '@/store/app'

const App: React.FC = () => {
  const location = useLocation()
  const setCurrentPath = useAppStore(state => state.setCurrentPath)
  React.useEffect(() => {
    setCurrentPath(location.pathname)
  }, [location.pathname])

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<RequireAuth><MainLayout /></RequireAuth>}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="monitor" element={<Monitor />} />
        <Route path="route-plan" element={<RoutePlan />} />
        <Route path="fatigue-alarms" element={<FatigueAlarms />} />
        <Route path="vehicles" element={<Vehicles />} />
        <Route path="drivers" element={<Drivers />} />
        <Route path="waybills" element={<Waybills />} />
        <Route path="escort" element={<Escort />} />
        <Route path="rescue" element={<Rescue />} />
        <Route path="weather" element={<Weather />} />
        <Route path="blockchain" element={<Blockchain />} />
        <Route path="profile" element={<Profile />} />
      </Route>
      <Route path="/big-screen" element={<DashboardLayout />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}

export default App
