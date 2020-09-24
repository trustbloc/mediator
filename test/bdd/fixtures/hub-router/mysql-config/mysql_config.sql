/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

\! echo "Configuring MySQL users...";

/*
Hub Router
*/
CREATE USER 'hubrouter'@'%' IDENTIFIED BY 'hubrouter-secret-pw';
GRANT ALL PRIVILEGES ON `hubrouter\_%` . * TO 'hubrouter'@'%';

/*
Aries Agent (mock wallet)
*/
CREATE USER 'aries'@'%' IDENTIFIED BY 'aries-secret-pw';
GRANT ALL PRIVILEGES ON * . * TO 'aries'@'%';
