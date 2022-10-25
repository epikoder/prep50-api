var accessToken = ''

const authStat = document.getElementById('stat-auth')

const authenticate = async () => {
    const user = document.getElementById('user').value
    const password = document.getElementById('password').value

    console.log(user, password)
    const res = await fetch('/auth/login', {
        method: 'POST',
        body: JSON.stringify({
            user, password,
            device_id: "a6b17bae-85f5-4e67-8cc1-cf21b308b97e",
            device_name: "NrYDrQMtc1l0"
        })
    })
    if (res.status === 200) {
        const data = await res.json()
        if (data.status === 'success') {
            accessToken = data.data.access
            authStat.style.color = 'green'
            return
        }
    }
    authStat.style.color = 'red'
}

const onPay = () => {
    const type = document.getElementById('type').value
    const action = document.getElementById('action').value


    const amount = type == 'waec' ? 10000 : (type == 'mock' ? 1000 : 5000)

    const paystack = new PaystackPop();
    paystack.newTransaction({
        key: 'pk_test_af4e0ef7bfdedcb27859289c27fd2a97100f1ba1',
        email: 'testuser@gmail.com',
        amount: amount * 100,
        metadata: {
            type: type,
            action: action,
        },
        callback: (response) => {
            document.getElementById('tx').innerText = response.reference
            if (response.status !== 'success') return alert('Transaction Failed')
            if (document.getElementById('shouldVerify').checked) {
                return verifyTranx(response.reference, type)
            }
        }
    });
}

/**
 * 
 * @param {string} reference
 * @param {string} type
 */
const verifyTranx = async (reference, type) => {
    const res = await fetch('/pay-verify', {
        method: 'POST',
        body: JSON.stringify({
            reference,
            type
        }),
        headers: {
            authorization: 'Bearer ' + accessToken
        }
    })
    if (res.status === 200) {
        const data = await res.json()
        if (data.status === 'success') {
            accessToken = data.data.access
            return
        }
    }
}

document.getElementById("pay-form").addEventListener('submit', (e) => e.preventDefault())
document.getElementById("login-form").addEventListener('submit', (e) => e.preventDefault())