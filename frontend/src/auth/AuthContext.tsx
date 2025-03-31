import React, { createContext, useEffect, useState } from "react";
import { submitFormData } from "../ui/forms/FormUtilities";

interface AuthData {
  username: string
}

interface RegisterData {
  username: string
  password: string
  email?: string
  rememberMe: boolean
}

export interface RegisterFormValidationErrors {
  username: string
  password: string
  email: string
}

interface LoginData {
  username: string
  password: string
  rememberMe: boolean
}

export interface LoginFormValidationErrors {
  username: string
  password: string
}

export interface AuthContextType {
  isLoading: boolean,
  isLoggedIn:  boolean,
  authData: AuthData,
  register (data: RegisterData, callback?: (success: boolean, responseData?: RegisterFormValidationErrors) => void): void,
  login (data: LoginData, callback?: (success: boolean, responseData?: LoginFormValidationErrors) => void): void,
  logout(callback?: (success: boolean) => void) :void,
}

const DEFAULT_AUTH_DATA: AuthData = {
  username: "",
}

export const AuthContext = createContext<AuthContextType>({
  isLoading: true,
  isLoggedIn: false,
  authData: DEFAULT_AUTH_DATA,
  register: () => {},
  login: () => {},
  logout: () => {}
});

async function register(setIsLoggedIn: React.Dispatch<React.SetStateAction<boolean>>, setAuthData: React.Dispatch<React.SetStateAction<AuthData>>, data: RegisterData, callback?: (success: boolean, responseData?: AuthData) => void) {
  const url = import.meta.env.VITE_API_REGISTER_URL
  const options: RequestInit = {
    credentials: "include",
    body: JSON.stringify({
      username: data.username,
      password: data.password,
      email: data.email || "",
      rememberMe: data.rememberMe,
    })
  }

  const registerCallback = (success: boolean, responseData: unknown) => {
    if (success) {
      setIsLoggedIn(true)
      setAuthData(responseData as AuthData)
    }
    if (callback !== undefined) {
      callback(success, responseData as RegisterFormValidationErrors)
    }
  }

  submitFormData(url, options, registerCallback)
}

async function login(setIsLoggedIn: React.Dispatch<React.SetStateAction<boolean>>, setAuthData: React.Dispatch<React.SetStateAction<AuthData>>, data: LoginData, callback?: (success: boolean, responseData?: AuthData) => void) {
  const url = import.meta.env.VITE_API_LOGIN_URL
  const options: RequestInit = {
    credentials: "include",
    body: JSON.stringify({
      username: data.username,
      password: data.password,
      rememberMe: data.rememberMe,
    })
  }

  const loginCallback = (success: boolean, responseData: unknown) => {
    if (success) {
      setIsLoggedIn(true)
      setAuthData(responseData as AuthData)
    }
    if (callback !== undefined) {
      callback(success, responseData as RegisterFormValidationErrors)
    }
  }

  submitFormData(url, options, loginCallback)
}

async function logout(setIsLoggedIn: React.Dispatch<React.SetStateAction<boolean>>, setAuthData: React.Dispatch<React.SetStateAction<AuthData>>, callback?: (success: boolean) => void) {
  const url = import.meta.env.VITE_API_LOGOUT_URL
  const options: RequestInit = {
    credentials: "include",
  }

  const logoutCallback = (success: boolean) => {
    if (success) {
      setIsLoggedIn(false)
      setAuthData(DEFAULT_AUTH_DATA)
    }
    if (callback !== undefined) {
      callback(false)
    }
  }

  submitFormData(url, options, logoutCallback)
}

async function setLoginStatus(
  setIsLoggedIn: React.Dispatch<React.SetStateAction<boolean>>,
  setAuthData: React.Dispatch<React.SetStateAction<AuthData>>,
  setIsLoading: React.Dispatch<React.SetStateAction<boolean>>,
) {
  const url = import.meta.env.VITE_API_VALIDATE_SESSION_URL

  try {
    const response = await fetch(url, {
      method: "POST",
      credentials: "include",
      signal: AbortSignal.timeout(5000),
    })

    console.log("TEST")
    console.log(response)
    
    if (response.ok) {
      const data = await response.json()
      setIsLoggedIn(true)
      setAuthData(data)
    }
  } catch (e) {
    console.error(e)
  } finally {
    setIsLoading(false)
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [authData, setAuthData] = useState<AuthData>(DEFAULT_AUTH_DATA)

  const registerClosure = (data: RegisterData, callback: (success: boolean) => void) => {
    register(setIsLoggedIn, setAuthData, data, callback)
  }

  const loginClosure = (data: LoginData, callback: (success: boolean) => void) => {
    login(setIsLoggedIn, setAuthData, data, callback)
  }

  const logoutClosure = (callback: (success: boolean) => void) => {
    logout(setIsLoggedIn, setAuthData, callback)
  }

  useEffect(() => {
    setLoginStatus(setIsLoggedIn, setAuthData, setIsLoading)
  }, [])

  return (
    <AuthContext.Provider value={{isLoading, isLoggedIn, authData, register: registerClosure, login: loginClosure, logout: logoutClosure}}>
      {children}
    </AuthContext.Provider>
  )
}