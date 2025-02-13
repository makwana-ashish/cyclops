import React, {useEffect, useState} from 'react';
import {Col, Divider, Row, Table, Alert} from 'antd';
import axios from 'axios';

interface Props {
    name: string;
    namespace: string;
}

interface port {
    name: string,
    protocol: string,
    port: string,
    targetPort: string,
}

interface service {
    ports: port[],
}

const Service = ({name, namespace}: Props) => {
    const [service, setService] = useState<service>({
        ports: []
    });
    const [error, setError] = useState({
        message: "",
        description: "",
    });

    function fetchService() {
        axios.get(window.__RUNTIME_CONFIG__.REACT_APP_CYCLOPS_CTRL_HOST + `/resources`,{
            params: {
                group: ``,
                version: `v1`,
                kind: `Service`,
                name: name,
                namespace: namespace
            }
        }).then(res => {
            setService(res.data)
        }).catch(error => {
            console.log(error)
            if (error.response === undefined) {
                setError({
                    message: String(error),
                    description: "Check if Cyclops backend is available on: " + window.__RUNTIME_CONFIG__.REACT_APP_CYCLOPS_CTRL_HOST
                })
            } else {
                setError(error.response.data);
            }
        })
    }

    useEffect(() => {
        fetchService()
        const interval = setInterval(() => fetchService(), 15000)
        return () => {
            clearInterval(interval);
        }
    }, []);

    return (
        <div>
            {
                error.message.length !== 0 && <Alert
                    message={error.message}
                    description={error.description}
                    type="error"
                    closable
                    afterClose={() => {setError({
                        message: "",
                        description: "",
                    })}}
                    style={{marginBottom: '20px'}}
                />
            }
            <Divider/>
            <Row>
                <Col span={24} style={{overflowX: "auto"}}>
                    <Table dataSource={service.ports}>
                        <Table.Column
                            title='Name'
                            dataIndex='name'
                            key='name'
                        />
                        <Table.Column
                            title='Protocol'
                            dataIndex='protocol'
                        />
                        <Table.Column
                            title='Port'
                            dataIndex='port'
                        />
                        <Table.Column
                            title='Target port'
                            dataIndex='targetPort'
                        />
                    </Table>
                </Col>
            </Row>
        </div>
    );
}

export default Service;
