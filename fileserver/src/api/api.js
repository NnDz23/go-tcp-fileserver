// export const fetchAPI = (url,requestOptions)=>{
//     console.log('fetch')
//     let res = {}
//     fetch(url,requestOptions)
//     .then((response) => res = response.json())
//     .catch((error) => res = {error:true,message:error.message})
//     console.log('return')
//     return res
// }

export const fetchWrapper = {
    get,
    post
};

function get(url) {
    const requestOptions = {
        method: 'GET'
    };
    return fetch(url, requestOptions).then(handleResponse).catch(handleError);
};

function post(url, body) {
    const requestOptions = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
    };
    return fetch(url, requestOptions).then(handleResponse).catch(handleError);
};

function handleResponse(response) {
    return response.json();
};

function handleError(error) {
    return Promise.reject(error);
};