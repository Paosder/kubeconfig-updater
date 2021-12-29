import dayjs from 'dayjs'
import { observer } from 'mobx-react-lite'
import { useResolve } from '../hooks/container'
import { useAutorun, useReaction } from '../hooks/mobx'
import browserLogger from '../logger/browserLogger'
import SnackbarStore from '../store/snackbarStore'
import { useContext } from './clusterRegisterRequester'

// TODO: replace this to store
export default observer(function RegisterErrorToNoti() {
  const snackbarStore = useResolve(SnackbarStore)
  const registerRequester = useContext()

  useReaction(
    () => registerRequester.error,
    (err) => {
      browserLogger.debug('got register error event')
      snackbarStore.push({
        key: dayjs().toString(),
        message: String(err),
        options: { variant: 'error' },
      })
    }
  )

  return <></>
})