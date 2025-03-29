import { useContext, useEffect, useState } from "react"
import { AuthContext } from "./AuthContext"
import { useNavigate } from "react-router-dom"
import { LoaderCircle } from "lucide-react"
import { FormError } from "../FormError"

interface AccountSettings {

}

async function fetchAccountSettings(username: string, signal: AbortSignal) {
    return null
}

function PasswordChange({ accountSettings }: { accountSettings: AccountSettings }) {
    return (
        <div>

        </div>
    )
}

async function handleFormSubmit(auth: AuthContextType, formData: FormData, navigate: NavigateFunction, setLoading: React.Dispatch<React.SetStateAction<boolean>>, setValidationErrors: React.Dispatch<React.SetStateAction<LoginFormValidationErrors>>) {
  setLoading(true)
  const redirectUrl = formData.get("referrer") as string || "/"
  const loginData = {
    username: formData.get("username") as string,
    password: formData.get("password") as string,
    rememberMe: formData.get("rememberMe") == "true" ? true : false,
  }

  console.log(loginData)

  const loginCallback = (success: boolean, responseData: LoginFormValidationErrors | undefined) => {
    if (success) {
      navigate(redirectUrl)
    } else {
      if (responseData != undefined) {
        setValidationErrors(responseData)
      }
    }
  }

  auth.login(loginData, loginCallback)
  setLoading(false)
}

function EmailChange({ accountSettings }: { accountSettings: AccountSettings }) {
  const [loading, setLoading] = useState(false)
  const [email, setEmail] = useState("")
  const navigate = useNavigate()
  const [validationErrors, setValidationErrors] = useState({
    email: "",
  })
  const auth = useContext(AuthContext)
  return (
    <form action={(formData) => {if (!loading) {handleFormSubmit(auth, formData, navigate, setLoading, setValidationErrors)}}}>
    <div className='formGroup'>
        <label htmlFor="username">Username</label>
        <input name="username" type="text" required={true} value={email} onChange={(event) => setEmail(event.target.value)}/>
        <FormError errorMessage={validationErrors.email} />
    </div>
    <button className='signInButton'>SIGN IN</button>
    </form>
  )
}

export function AccountSettingsPage() {
    const auth = useContext(AuthContext)
    const navigate = useNavigate()
    const [loadingAccountSettings, setLoadingAccountSettings] = useState(true)
    const [accountSettingsData, setAccountSettingsData] = useState<AccountSettings | null>(null)

  useEffect(() => {
    let ignore = false
    setLoadingAccountSettings(true)
    const controller = new AbortController()
    const signal = controller.signal;

    (async() => {
      const accountSettingsData = fetchAccountSettings(auth.authData.username, signal)
      if (!ignore) {
        setAccountSettingsData(await accountSettingsData)
        setLoadingAccountSettings(false)
      }}
    )()

    return () => {
      ignore = true
      controller.abort("page change")
    }
  }, [])

    if (!auth.isLoggedIn) {
        navigate("/")
    }

    if (loadingAccountSettings) {
        return (
            <div>
                <LoaderCircle className="loaderSpin"/>
            </div>
        )
    }

    if (accountSettingsData === null) {
        return (
            <div>
                Error getting data.
            </div>
        )
    }

    return (
        <div>
            <PasswordChange accountSettings={accountSettingsData}/>
            <EmailChange accountSettings={accountSettingsData}/>
        </div>
    )
}